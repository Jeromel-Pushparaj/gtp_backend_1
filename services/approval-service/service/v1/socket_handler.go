package v1

import (
	"log"

	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type SocketHandler struct {
	client          *socketmode.Client
	approvalService *ApprovalService
}

func NewSocketHandler(botToken, appToken string, approvalService *ApprovalService) *SocketHandler {
	api := slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
		slack.OptionDebug(true),
	)

	client := socketmode.New(api)

	return &SocketHandler{
		client:          client,
		approvalService: approvalService,
	}
}

func (sh *SocketHandler) Start() {
	go func() {
		for evt := range sh.client.Events {
			switch evt.Type {
			case socketmode.EventTypeInteractive:
				sh.handleInteractive(evt)
			case socketmode.EventTypeConnecting:
				log.Println("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				log.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				log.Println("Connected to Slack with Socket Mode.")
			default:
				log.Printf("Unexpected event type received: %s\n", evt.Type)
			}
		}
	}()

	log.Println("Starting Socket Mode client...")
	if err := sh.client.Run(); err != nil {
		log.Fatalf("Socket Mode client error: %v", err)
	}
}

func (sh *SocketHandler) handleInteractive(evt socketmode.Event) {
	callback, ok := evt.Data.(slack.InteractionCallback)
	if !ok {
		log.Printf("Ignored %+v\n", evt)
		sh.client.Ack(*evt.Request)
		return
	}

	log.Printf("Interaction received: %+v\n", callback)

	if callback.Type != slack.InteractionTypeBlockActions {
		sh.client.Ack(*evt.Request)
		return
	}

	if len(callback.ActionCallback.BlockActions) == 0 {
		sh.client.Ack(*evt.Request)
		return
	}

	action := callback.ActionCallback.BlockActions[0]
	requestID := action.Value
	userID := callback.User.ID
	userName := callback.User.Name

	var approved bool
	var reason string

	switch action.ActionID {
	case constants.ActionApprove:
		approved = true
		reason = "Approved by manager"
	case constants.ActionReject:
		approved = false
		reason = "Rejected by manager"
	default:
		sh.client.Ack(*evt.Request)
		return
	}

	err := sh.approvalService.HandleApproval(requestID, approved, userID, userName, reason)
	if err != nil {
		log.Printf("Error handling approval: %v", err)
		sh.client.Ack(*evt.Request, map[string]interface{}{
			"text": "Error processing approval: " + err.Error(),
		})
		return
	}

	statusText := "rejected"
	if approved {
		statusText = "approved"
	}

	sh.client.Ack(*evt.Request, map[string]interface{}{
		"text": "Request " + statusText + " successfully",
	})

	log.Printf("Approval %s: request_id=%s, user=%s", statusText, requestID, userName)
}
