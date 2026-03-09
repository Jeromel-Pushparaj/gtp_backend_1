package v1

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/db"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/models"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/resources"
	"github.com/slack-go/slack"
)

type ApprovalService struct {
	repo         *db.ApprovalRepository
	slackService *SlackService
	kafkaService *KafkaService
}

func NewApprovalService(repo *db.ApprovalRepository, slackService *SlackService, kafkaService *KafkaService) *ApprovalService {
	return &ApprovalService{
		repo:         repo,
		slackService: slackService,
		kafkaService: kafkaService,
	}
}

func (s *ApprovalService) ProcessApprovalRequest(msg *resources.ApprovalRequestMessage) error {
	requestID := msg.RequestID
	if requestID == "" {
		requestID = uuid.New().String()
	}

	log.Printf("=== KAFKA CONSUMER: Processing approval request ===")
	log.Printf("RequestID=%s, RequesterID=%s, ApproverID=%s", requestID, msg.RequesterID, msg.ApproverID)

	requestDataJSON, err := json.Marshal(msg.RequestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	approval := &models.ApprovalRequest{
		RequestID:     requestID,
		RequesterID:   msg.RequesterID,
		RequesterName: msg.RequesterName,
		ApproverID:    msg.ApproverID,
		ApproverName:  msg.ApproverName,
		ChannelID:     msg.ChannelID,
		RequestType:   msg.RequestType,
		RequestData:   string(requestDataJSON),
		Status:        constants.ApprovalStatusPending,
		Title:         msg.Title,
		Description:   msg.Description,
		Priority:      msg.Priority,
		Category:      msg.Category,
		Attachments:   msg.Attachments,
		DueDate:       msg.DueDate,
	}

	if err := s.repo.Create(approval); err != nil {
		log.Printf("ERROR: Failed to create approval record in DB: %v", err)
		return fmt.Errorf("failed to create approval record: %w", err)
	}

	log.Printf("✓ SUCCESS: Created approval request in database: RequestID=%s", requestID)

	var timestamp string

	isAppDM := false
	if msg.RequestData != nil {
		if useAppDM, ok := msg.RequestData["use_app_dm"].(bool); ok {
			isAppDM = useAppDM
		}
	}

	if isAppDM {
		timestamp, err = s.sendSlackApprovalMessageToAppDM(msg, msg.ChannelID)
	} else {
		timestamp, err = s.sendSlackApprovalMessage(msg)
	}
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	// timestamp, err := s.sendSlackApprovalMessage(msg)
	// if err != nil {
	// 	return fmt.Errorf("failed to send slack message: %w", err)
	// }

	approval.MessageTS = timestamp
	s.repo.UpdateMessageTS(requestID, timestamp)

	log.Printf("Sent Slack approval message for request: %s to channel: %s", requestID, msg.ChannelID)
	return nil
}

func (s *ApprovalService) sendSlackApprovalMessage(msg *resources.ApprovalRequestMessage) (string, error) {
	headerText := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Approval Request from %s*", msg.RequesterName), false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	messageText := slack.NewTextBlockObject(slack.MarkdownType, msg.Message, false, false)
	messageSection := slack.NewSectionBlock(messageText, nil, nil)

	divider := slack.NewDividerBlock()

	approveButton := slack.NewButtonBlockElement(
		constants.ActionApprove,
		msg.RequestID,
		slack.NewTextBlockObject(slack.PlainTextType, "Approve", false, false),
	)
	approveButton.Style = slack.StylePrimary

	rejectButton := slack.NewButtonBlockElement(
		constants.ActionReject,
		msg.RequestID,
		slack.NewTextBlockObject(slack.PlainTextType, "Reject", false, false),
	)
	rejectButton.Style = slack.StyleDanger

	actionBlock := slack.NewActionBlock(
		"approval_actions",
		approveButton,
		rejectButton,
	)

	blocks := []slack.Block{
		headerSection,
		messageSection,
		divider,
		actionBlock,
	}

	timestamp, err := s.slackService.SendBlockMessage(
		msg.ChannelID,
		blocks,
		fmt.Sprintf("Approval request from %s", msg.RequesterName),
	)

	return timestamp, err
}

func (s *ApprovalService) HandleApproval(requestID string, approved bool, userID string, userName string, reason string, approverComment string) error {

	var approval *models.ApprovalRequest
	var err error
	maxRetries := 10
	retryDelay := 1 * time.Second

	log.Printf("HandleApproval called for request: %s", requestID)

	for i := 0; i < maxRetries; i++ {
		approval, err = s.repo.GetByRequestID(requestID)
		if err == nil {
			log.Printf("Approval request %s found on attempt %d", requestID, i+1)
			break
		}
		if i < maxRetries-1 {
			log.Printf("Approval request %s not found yet (attempt %d/%d), retrying in %v...", requestID, i+1, maxRetries, retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Printf("ERROR: Approval request %s not found after %d retries: %v", requestID, maxRetries, err)
		return fmt.Errorf("%s: %w", constants.ErrorApprovalNotFound, err)
	}

	if approval.Status != constants.ApprovalStatusPending {
		return fmt.Errorf(constants.ErrorApprovalAlreadyProcessed)
	}

	status := constants.ApprovalStatusRejected
	if approved {
		status = constants.ApprovalStatusApproved
	}

	if err := s.repo.UpdateStatus(requestID, status, approved, userID, reason, approverComment); err != nil {
		return fmt.Errorf("failed to update approval status: %w", err)
	}

	log.Printf("Approval %s: status=%s, by=%s, comment=%s", requestID, status, userName, approverComment)

	if err := s.updateSlackMessage(approval, approved, userName, approverComment); err != nil {
		log.Printf("Warning: failed to update slack message: %v", err)
	}

	if err := s.notifyRequester(approval, approved, userName, approverComment); err != nil {
		log.Printf("Warning: failed to notify requester: %v", err)
	}

	var requestData map[string]interface{}
	if err := json.Unmarshal([]byte(approval.RequestData), &requestData); err != nil {
		log.Printf("Warning: failed to unmarshal request data: %v", err)
		requestData = make(map[string]interface{})
	}

	completedMsg := &resources.ApprovalCompletedMessage{
		RequestID:       requestID,
		Status:          status,
		Approved:        approved,
		ProcessedBy:     userID,
		ProcessedAt:     time.Now(),
		Reason:          reason,
		ApproverComment: approverComment,
		RequestData:     requestData,
	}

	if err := s.kafkaService.Publish(constants.KafkaTopicApprovalCompleted, requestID, completedMsg); err != nil {
		return fmt.Errorf("failed to publish to kafka: %w", err)
	}

	log.Printf("Published approval completion to Kafka: %s", requestID)
	return nil
}

func (s *ApprovalService) updateSlackMessage(approval *models.ApprovalRequest, approved bool, userName string, approverComment string) error {
	statusText := "REJECTED"
	statusEmoji := ":x:"
	if approved {
		statusText = "APPROVED"
		statusEmoji = ":white_check_mark:"
	}

	var headerText string
	if approval.Title != "" {
		headerText = fmt.Sprintf("*:memo: %s*", approval.Title)
	} else {
		headerText = fmt.Sprintf("*Approval Request from %s*", approval.RequesterName)
	}

	headerSection := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false),
		nil,
		nil,
	)

	var messageText strings.Builder
	messageText.WriteString(fmt.Sprintf("*Status:* %s *%s*\n", statusEmoji, statusText))
	messageText.WriteString(fmt.Sprintf("*Processed by:* %s", userName))

	if approval.Category != "" {
		messageText.WriteString(fmt.Sprintf("\n*Category:* %s", strings.Title(approval.Category)))
	}
	if approval.Priority != "" {
		messageText.WriteString(fmt.Sprintf("\n*Priority:* %s", strings.Title(approval.Priority)))
	}

	if approverComment != "" {
		messageText.WriteString(fmt.Sprintf("\n*Comment:* _%s_", approverComment))
	}

	messageSection := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, messageText.String(), false, false),
		nil,
		nil,
	)

	blocks := []slack.Block{
		headerSection,
		messageSection,
	}

	return s.slackService.UpdateMessage(
		approval.ChannelID,
		approval.MessageTS,
		blocks,
	)
}

func (s *ApprovalService) CreateApprovalFromSlashCommand(
	channelID string,
	requesterID string,
	approverID string,
	title string,
	description string,
	priority string,
	category string,
	attachments string,
	dueDate *time.Time,
) error {
	requester, err := s.slackService.GetUserByID(requesterID)
	if err != nil {
		return fmt.Errorf("failed to get requester info: %w", err)
	}

	approver, err := s.slackService.GetUserByID(approverID)
	if err != nil {
		return fmt.Errorf("failed to get approver info: %w", err)
	}

	requestID := uuid.New().String()

	requestData := map[string]interface{}{
		"title":       title,
		"description": description,
		"priority":    priority,
		"category":    category,
		"attachments": attachments,
	}
	if dueDate != nil {
		requestData["due_date"] = dueDate.Format("2006-01-02")
	}

	requestDataJSON, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	approval := &models.ApprovalRequest{
		RequestID:     requestID,
		RequesterID:   requester.ID,
		RequesterName: requester.RealName,
		ApproverID:    approver.ID,
		ApproverName:  approver.RealName,
		ChannelID:     channelID,
		RequestType:   "slash_command",
		RequestData:   string(requestDataJSON),
		Status:        constants.ApprovalStatusPending,
		Title:         title,
		Description:   description,
		Priority:      priority,
		Category:      category,
		Attachments:   attachments,
		DueDate:       dueDate,
	}

	if err := s.repo.Create(approval); err != nil {
		return fmt.Errorf("failed to create approval record: %w", err)
	}

	log.Printf("Created approval request via slash command: %s", requestID)

	timestamp, err := s.sendRichSlackApprovalMessage(approval)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	approval.MessageTS = timestamp
	s.repo.UpdateMessageTS(requestID, timestamp)

	log.Printf("Sent rich Slack approval message for request: %s", requestID)
	return nil
}

func (s *ApprovalService) sendRichSlackApprovalMessage(approval *models.ApprovalRequest) (string, error) {
	headerText := slack.NewTextBlockObject(
		slack.MarkdownType,
		fmt.Sprintf("*:memo: %s*", approval.Title),
		false,
		false,
	)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	priorityEmoji := ":small_blue_diamond:"
	switch approval.Priority {
	case constants.PriorityUrgent:
		priorityEmoji = ":red_circle:"
	case constants.PriorityHigh:
		priorityEmoji = ":large_orange_diamond:"
	case constants.PriorityMedium:
		priorityEmoji = ":large_yellow_circle:"
	case constants.PriorityLow:
		priorityEmoji = ":white_circle:"
	}

	var detailsText strings.Builder
	detailsText.WriteString(fmt.Sprintf("*Requested by:* <@%s>\n", approval.RequesterID))
	detailsText.WriteString(fmt.Sprintf("*Approver:* <@%s>\n", approval.ApproverID))
	detailsText.WriteString(fmt.Sprintf("*Category:* %s\n", strings.Title(approval.Category)))
	detailsText.WriteString(fmt.Sprintf("*Priority:* %s %s", priorityEmoji, strings.Title(approval.Priority)))

	if approval.DueDate != nil {
		detailsText.WriteString(fmt.Sprintf("\n*Due Date:* %s", approval.DueDate.Format("Jan 02, 2006")))
	}

	detailsSection := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, detailsText.String(), false, false),
		nil,
		nil,
	)

	descriptionSection := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Description:*\n%s", approval.Description), false, false),
		nil,
		nil,
	)

	blocks := []slack.Block{
		headerSection,
		detailsSection,
		descriptionSection,
	}

	if approval.Attachments != "" {
		attachmentsSection := slack.NewSectionBlock(
			slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Attachments:*\n%s", approval.Attachments), false, false),
			nil,
			nil,
		)
		blocks = append(blocks, attachmentsSection)
	}

	divider := slack.NewDividerBlock()
	blocks = append(blocks, divider)

	approveButton := slack.NewButtonBlockElement(
		constants.ActionApprove,
		approval.RequestID,
		slack.NewTextBlockObject(slack.PlainTextType, "Approve", false, false),
	)
	approveButton.Style = slack.StylePrimary

	rejectButton := slack.NewButtonBlockElement(
		constants.ActionReject,
		approval.RequestID,
		slack.NewTextBlockObject(slack.PlainTextType, "Reject", false, false),
	)
	rejectButton.Style = slack.StyleDanger

	actionBlock := slack.NewActionBlock(
		"approval_actions",
		approveButton,
		rejectButton,
	)
	blocks = append(blocks, actionBlock)

	timestamp, err := s.slackService.SendBlockMessage(
		approval.ChannelID,
		blocks,
		fmt.Sprintf("Approval request: %s", approval.Title),
	)

	return timestamp, err
}

func (s *ApprovalService) sendSlackApprovalMessageToAppDM(msg *resources.ApprovalRequestMessage, dmChannelID string) (string, error) {

	var headerText strings.Builder
	if msg.Title != "" {
		headerText.WriteString(fmt.Sprintf("*:memo: %s*\n\n", msg.Title))
	} else {
		headerText.WriteString(fmt.Sprintf("*Approval Request from <@%s>*\n\n", msg.RequesterID))
	}

	if msg.Message != "" {
		headerText.WriteString(fmt.Sprintf("%s\n\n", msg.Message))
	}

	var detailsText strings.Builder
	detailsText.WriteString("*Request Details:*\n")

	if msg.RequestData != nil {
		if oldDomain, ok := msg.RequestData["old_domain_name"].(string); ok {
			detailsText.WriteString(fmt.Sprintf("• *Old Domain Name:* `%s`\n", oldDomain))
		}
		if newDomain, ok := msg.RequestData["new_domain_name"].(string); ok {
			detailsText.WriteString(fmt.Sprintf("• *New Domain Name:* `%s`\n", newDomain))
		}
		if changeReason, ok := msg.RequestData["change_reason"].(string); ok && changeReason != "" {
			detailsText.WriteString(fmt.Sprintf("• *Change Reason:* %s\n", changeReason))
		}
		if additionalInfo, ok := msg.RequestData["additional_info"].(string); ok && additionalInfo != "" {
			detailsText.WriteString(fmt.Sprintf("• *Additional Info:* %s\n", additionalInfo))
		}
	}

	if msg.Priority != "" {
		priorityEmoji := ":small_blue_diamond:"
		switch msg.Priority {
		case constants.PriorityUrgent:
			priorityEmoji = ":red_circle:"
		case constants.PriorityHigh:
			priorityEmoji = ":large_orange_diamond:"
		case constants.PriorityMedium:
			priorityEmoji = ":large_yellow_circle:"
		case constants.PriorityLow:
			priorityEmoji = ":white_circle:"
		}
		detailsText.WriteString(fmt.Sprintf("• *Priority:* %s %s\n", priorityEmoji, strings.Title(msg.Priority)))
	}

	if msg.Category != "" {
		detailsText.WriteString(fmt.Sprintf("• *Category:* %s\n", strings.Title(msg.Category)))
	}

	headerText.WriteString(detailsText.String())

	headerText.WriteString(fmt.Sprintf("\n<@%s> - Please review and approve/reject this change.", msg.ApproverID))

	headerSection := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, headerText.String(), false, false),
		nil,
		nil,
	)

	divider := slack.NewDividerBlock()

	approveButton := slack.NewButtonBlockElement(
		constants.ActionApprove,
		msg.RequestID,
		slack.NewTextBlockObject(slack.PlainTextType, "Approve", false, false),
	)
	approveButton.Style = slack.StylePrimary

	rejectButton := slack.NewButtonBlockElement(
		constants.ActionReject,
		msg.RequestID,
		slack.NewTextBlockObject(slack.PlainTextType, "Reject", false, false),
	)
	rejectButton.Style = slack.StyleDanger

	actionBlock := slack.NewActionBlock(
		"approval_actions",
		approveButton,
		rejectButton,
	)

	blocks := []slack.Block{
		headerSection,
		divider,
		actionBlock,
	}

	timestamp, err := s.slackService.SendBlockMessage(
		dmChannelID,
		blocks,
		fmt.Sprintf("Approval request from %s", msg.RequesterName),
	)

	return timestamp, err
}

func (s *ApprovalService) notifyRequester(approval *models.ApprovalRequest, approved bool, approverName string, comment string) error {
	statusText := "REJECTED"
	statusEmoji := ":x:"
	if approved {
		statusText = "APPROVED"
		statusEmoji = ":white_check_mark:"
	}

	var requestData map[string]interface{}
	if err := json.Unmarshal([]byte(approval.RequestData), &requestData); err != nil {
		requestData = make(map[string]interface{})
	}

	var messageText strings.Builder
	messageText.WriteString(fmt.Sprintf("%s *Your request has been %s*\n\n", statusEmoji, statusText))
	messageText.WriteString(fmt.Sprintf("*Request ID:* `%s`\n", approval.RequestID))
	messageText.WriteString(fmt.Sprintf("*Approver:* <@%s>\n", approval.ApproverID))
	messageText.WriteString(fmt.Sprintf("*Decision:* %s\n", statusText))

	if oldDomain, ok := requestData["old_domain_name"].(string); ok {
		if newDomain, ok := requestData["new_domain_name"].(string); ok {
			messageText.WriteString(fmt.Sprintf("*Domain Change:* `%s` → `%s`\n", oldDomain, newDomain))
		}
	}

	if comment != "" {
		messageText.WriteString(fmt.Sprintf("\n*Comment from approver:*\n> %s\n", comment))
	}

	messageText.WriteString(fmt.Sprintf("\n_Processed at: %s_", time.Now().Format("2006-01-02 15:04:05")))

	headerSection := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, messageText.String(), false, false),
		nil,
		nil,
	)

	contextElements := []slack.MixedElement{
		slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Status:* %s %s", statusEmoji, statusText), false, false),
	}
	contextBlock := slack.NewContextBlock("", contextElements...)

	divider := slack.NewDividerBlock()

	blocks := []slack.Block{
		headerSection,
		divider,
		contextBlock,
	}

	// Try to open DM with requester
	dmChannelID, err := s.slackService.OpenDMChannel(approval.RequesterID)
	if err != nil {
		log.Printf("Failed to open DM with requester %s, trying channel: %v", approval.RequesterID, err)
		// Fallback to original channel if DM fails
		_, err = s.slackService.SendBlockMessage(
			approval.ChannelID,
			blocks,
			fmt.Sprintf("Notification for <@%s>: Your request has been %s", approval.RequesterID, statusText),
		)
		return err
	}

	_, err = s.slackService.SendBlockMessage(
		dmChannelID,
		blocks,
		fmt.Sprintf("Your approval request has been %s", statusText),
	)

	log.Printf("Sent notification to requester %s: request %s was %s", approval.RequesterName, approval.RequestID, statusText)
	return err
}
