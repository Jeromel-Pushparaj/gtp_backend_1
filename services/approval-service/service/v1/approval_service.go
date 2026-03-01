package v1

import (
	"encoding/json"
	"fmt"
	"log"
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

func (s *ApprovalService) HandleApproval(requestID string, approved bool, userID string, userName string, reason string) error {
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

	if err := s.repo.UpdateStatus(requestID, status, approved, userID, reason); err != nil {
		return fmt.Errorf("failed to update approval status: %w", err)
	}

	log.Printf("Approval %s: status=%s, by=%s", requestID, status, userName)

	if err := s.updateSlackMessage(approval, approved, userName); err != nil {
		log.Printf("Warning: failed to update slack message: %v", err)
	}

	var requestData map[string]interface{}
	if err := json.Unmarshal([]byte(approval.RequestData), &requestData); err != nil {
		log.Printf("Warning: failed to unmarshal request data: %v", err)
		requestData = make(map[string]interface{})
	}

	completedMsg := &resources.ApprovalCompletedMessage{
		RequestID:   requestID,
		Status:      status,
		Approved:    approved,
		ProcessedBy: userID,
		ProcessedAt: time.Now(),
		Reason:      reason,
		RequestData: requestData,
	}

	if err := s.kafkaService.Publish(constants.KafkaTopicApprovalCompleted, requestID, completedMsg); err != nil {
		return fmt.Errorf("failed to publish to kafka: %w", err)
	}

	log.Printf("Published approval completion to Kafka: %s", requestID)
	return nil
}

func (s *ApprovalService) updateSlackMessage(approval *models.ApprovalRequest, approved bool, userName string) error {
	statusText := "REJECTED"
	statusEmoji := ":x:"
	if approved {
		statusText = "APPROVED"
		statusEmoji = ":white_check_mark:"
	}

	headerText := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Approval Request from %s*", approval.RequesterName), false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	messageText := fmt.Sprintf("Request Type: %s\nStatus: %s *%s*\nProcessed by: %s",
		approval.RequestType, statusEmoji, statusText, userName)

	messageSection := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, messageText, false, false),
		nil, nil,
	)

	blocks := []slack.Block{
		headerSection,
		messageSection,
	}

	err := s.slackService.UpdateBlockMessage(
		approval.ChannelID,
		approval.MessageTS,
		blocks,
	)

	return err
}
