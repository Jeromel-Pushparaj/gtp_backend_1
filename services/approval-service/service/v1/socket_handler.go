package v1

import (
	"log"
	"strings"

	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type SocketHandler struct {
	client          *socketmode.Client
	api             *slack.Client
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
		api:             api,
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

	switch callback.Type {
	case slack.InteractionTypeBlockActions:
		sh.handleBlockAction(evt, callback)
	case slack.InteractionTypeViewSubmission:
		sh.handleModalSubmission(evt, callback)
	default:
		sh.client.Ack(*evt.Request)
	}
}

func (sh *SocketHandler) handleBlockAction(evt socketmode.Event, callback slack.InteractionCallback) {
	if len(callback.ActionCallback.BlockActions) == 0 {
		sh.client.Ack(*evt.Request)
		return
	}

	action := callback.ActionCallback.BlockActions[0]
	requestID := action.Value

	var modalView slack.ModalViewRequest
	var err error

	switch action.ActionID {
	case constants.ActionApprove:
		modalView, err = sh.buildApprovalModal(requestID, true)
	case constants.ActionReject:
		modalView, err = sh.buildApprovalModal(requestID, false)
	default:
		sh.client.Ack(*evt.Request)
		return
	}

	if err != nil {
		log.Printf("Error building modal: %v", err)
		sh.client.Ack(*evt.Request, map[string]interface{}{
			"text": "Error opening modal: " + err.Error(),
		})
		return
	}

	_, err = sh.api.OpenView(callback.TriggerID, modalView)
	if err != nil {
		log.Printf("Error opening modal: %v", err)
		sh.client.Ack(*evt.Request, map[string]interface{}{
			"text": "Error opening modal: " + err.Error(),
		})
		return
	}

	sh.client.Ack(*evt.Request)
	log.Printf("Modal opened for request: %s", requestID)
}

func (sh *SocketHandler) buildApprovalModal(requestID string, isApprove bool) (slack.ModalViewRequest, error) {
	var title, submit, placeholder, callbackID string

	if isApprove {
		title = constants.ModalTitleApprove
		submit = constants.ModalSubmitApprove
		placeholder = constants.ModalPlaceholderApprove
		callbackID = constants.ModalCallbackApprove
	} else {
		title = constants.ModalTitleReject
		submit = constants.ModalSubmitReject
		placeholder = constants.ModalPlaceholderReject
		callbackID = constants.ModalCallbackReject
	}

	commentInput := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, placeholder, false, false),
		constants.InputActionComment,
	)
	commentInput.Multiline = true
	commentInput.MaxLength = constants.CommentMaxLength

	commentBlock := slack.NewInputBlock(
		constants.InputBlockComment,
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalLabelComment, false, false),
		nil,
		commentInput,
	)
	commentBlock.Optional = isApprove

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			commentBlock,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:            slack.ViewType("modal"),
		Title:           slack.NewTextBlockObject(slack.PlainTextType, title, false, false),
		Close:           slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false),
		Submit:          slack.NewTextBlockObject(slack.PlainTextType, submit, false, false),
		Blocks:          blocks,
		CallbackID:      callbackID,
		PrivateMetadata: requestID,
	}

	return modalRequest, nil
}

func (sh *SocketHandler) handleModalSubmission(evt socketmode.Event, callback slack.InteractionCallback) {
	requestID := callback.View.PrivateMetadata
	callbackID := callback.View.CallbackID
	userID := callback.User.ID
	userName := callback.User.Name

	commentValue := ""
	if inputBlock, ok := callback.View.State.Values[constants.InputBlockComment]; ok {
		if inputAction, ok := inputBlock[constants.InputActionComment]; ok {
			commentValue = strings.TrimSpace(inputAction.Value)
		}
	}

	var approved bool
	var reason string

	switch callbackID {
	case constants.ModalCallbackApprove:
		approved = true
		reason = "Approved by manager"
	case constants.ModalCallbackReject:
		approved = false
		reason = "Rejected by manager"
		if commentValue == "" {
			sh.client.Ack(*evt.Request, map[string]interface{}{
				"response_action": "errors",
				"errors": map[string]string{
					constants.InputBlockComment: constants.ErrorCommentRequired,
				},
			})
			return
		}
	default:
		sh.client.Ack(*evt.Request)
		return
	}

	err := sh.approvalService.HandleApproval(requestID, approved, userID, userName, reason, commentValue)
	if err != nil {
		log.Printf("Error handling approval: %v", err)
		sh.client.Ack(*evt.Request, map[string]interface{}{
			"response_action": "errors",
			"errors": map[string]string{
				constants.InputBlockComment: "Error processing approval: " + err.Error(),
			},
		})
		return
	}

	sh.client.Ack(*evt.Request)

	statusText := "rejected"
	if approved {
		statusText = "approved"
	}

	log.Printf("Approval %s: request_id=%s, user=%s, comment=%s", statusText, requestID, userName, commentValue)
}
