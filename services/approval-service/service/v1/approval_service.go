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
	}

	if err := s.repo.Create(approval); err != nil {
		return fmt.Errorf("failed to create approval record: %w", err)
	}

	log.Printf("Created approval request: %s", requestID)

	timestamp, err := s.sendSlackApprovalMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	approval.MessageTS = timestamp
	s.repo.UpdateMessageTS(requestID, timestamp)

	log.Printf("Sent Slack approval message for request: %s", requestID)
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
	approval, err := s.repo.GetByRequestID(requestID)
	if err != nil {
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
