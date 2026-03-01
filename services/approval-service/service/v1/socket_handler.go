package v1

import (
	"fmt"
	"log"
	"strings"
	"time"

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
			case socketmode.EventTypeSlashCommand:
				sh.handleSlashCommand(evt)
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
	case constants.ActionOpenApprovalForm:
		modalView, err = sh.buildCreateApprovalModal(callback.Channel.ID, callback.User.ID)
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

	if action.ActionID == constants.ActionOpenApprovalForm {
		log.Printf("Create approval modal opened via button for user: %s", callback.User.ID)
	} else {
		log.Printf("Modal opened for request: %s", requestID)
	}
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
	case constants.ModalCallbackCreateApproval:
		sh.handleCreateApprovalSubmission(evt, callback)
		return
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

func (sh *SocketHandler) handleSlashCommand(evt socketmode.Event) {
	command, ok := evt.Data.(slack.SlashCommand)
	if !ok {
		log.Printf("Ignored slash command event: %+v\n", evt)
		sh.client.Ack(*evt.Request)
		return
	}

	log.Printf("Slash command received: %s", command.Command)

	switch command.Command {
	case constants.SlashCommandApprovalRequest:
		sh.handleApprovalRequestCommand(evt, command)
	default:
		sh.client.Ack(*evt.Request)
	}
}

func (sh *SocketHandler) handleApprovalRequestCommand(evt socketmode.Event, command slack.SlashCommand) {
	modalView, err := sh.buildCreateApprovalModal(command.ChannelID, command.UserID)
	if err != nil {
		log.Printf("Error building create approval modal: %v", err)
		sh.client.Ack(*evt.Request, map[string]interface{}{
			"text": "Error opening form: " + err.Error(),
		})
		return
	}

	_, err = sh.api.OpenView(command.TriggerID, modalView)
	if err != nil {
		log.Printf("Error opening create approval modal: %v", err)
		sh.client.Ack(*evt.Request, map[string]interface{}{
			"text": "Error opening form: " + err.Error(),
		})
		return
	}

	sh.client.Ack(*evt.Request)
	log.Printf("Create approval modal opened for user: %s", command.UserID)
}

func (sh *SocketHandler) buildCreateApprovalModal(channelID, userID string) (slack.ModalViewRequest, error) {
	titleInput := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalPlaceholderTitle, false, false),
		constants.InputActionTitle,
	)
	titleInput.MaxLength = constants.TitleMaxLength

	titleBlock := slack.NewInputBlock(
		constants.InputBlockTitle,
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalLabelTitle, false, false),
		nil,
		titleInput,
	)

	descriptionInput := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalPlaceholderDescription, false, false),
		constants.InputActionDescription,
	)
	descriptionInput.Multiline = true
	descriptionInput.MaxLength = constants.DescriptionMaxLength

	descriptionBlock := slack.NewInputBlock(
		constants.InputBlockDescription,
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalLabelDescription, false, false),
		nil,
		descriptionInput,
	)

	approverSelect := slack.NewOptionsSelectBlockElement(
		slack.OptTypeUser,
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalPlaceholderApprover, false, false),
		constants.InputActionApprover,
	)

	approverBlock := slack.NewInputBlock(
		constants.InputBlockApprover,
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalLabelApprover, false, false),
		nil,
		approverSelect,
	)

	priorityOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject(constants.PriorityLow, slack.NewTextBlockObject(slack.PlainTextType, "Low", false, false), nil),
		slack.NewOptionBlockObject(constants.PriorityMedium, slack.NewTextBlockObject(slack.PlainTextType, "Medium", false, false), nil),
		slack.NewOptionBlockObject(constants.PriorityHigh, slack.NewTextBlockObject(slack.PlainTextType, "High", false, false), nil),
		slack.NewOptionBlockObject(constants.PriorityUrgent, slack.NewTextBlockObject(slack.PlainTextType, "Urgent", false, false), nil),
	}

	priorityRadio := slack.NewRadioButtonsBlockElement(
		constants.InputActionPriority,
		priorityOptions...,
	)
	priorityRadio.InitialOption = priorityOptions[1]

	priorityBlock := slack.NewInputBlock(
		constants.InputBlockPriority,
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalLabelPriority, false, false),
		nil,
		priorityRadio,
	)

	categoryOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject(constants.CategoryCodeChange, slack.NewTextBlockObject(slack.PlainTextType, "Code Change", false, false), nil),
		slack.NewOptionBlockObject(constants.CategoryMergeRequest, slack.NewTextBlockObject(slack.PlainTextType, "Merge Request", false, false), nil),
		slack.NewOptionBlockObject(constants.CategoryDeployment, slack.NewTextBlockObject(slack.PlainTextType, "Deployment", false, false), nil),
		slack.NewOptionBlockObject(constants.CategoryAccess, slack.NewTextBlockObject(slack.PlainTextType, "Access Request", false, false), nil),
		slack.NewOptionBlockObject(constants.CategoryInfrastructure, slack.NewTextBlockObject(slack.PlainTextType, "Infrastructure", false, false), nil),
		slack.NewOptionBlockObject(constants.CategoryOther, slack.NewTextBlockObject(slack.PlainTextType, "Other", false, false), nil),
	}

	categorySelect := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalPlaceholderCategory, false, false),
		constants.InputActionCategory,
		categoryOptions...,
	)

	categoryBlock := slack.NewInputBlock(
		constants.InputBlockCategory,
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalLabelCategory, false, false),
		nil,
		categorySelect,
	)

	attachmentsInput := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalPlaceholderAttachments, false, false),
		constants.InputActionAttachments,
	)
	attachmentsInput.MaxLength = constants.AttachmentsMaxLength

	attachmentsBlock := slack.NewInputBlock(
		constants.InputBlockAttachments,
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalLabelAttachments, false, false),
		nil,
		attachmentsInput,
	)
	attachmentsBlock.Optional = true

	dueDatePicker := slack.NewDatePickerBlockElement(constants.InputActionDueDate)

	dueDateBlock := slack.NewInputBlock(
		constants.InputBlockDueDate,
		slack.NewTextBlockObject(slack.PlainTextType, constants.ModalLabelDueDate, false, false),
		nil,
		dueDatePicker,
	)
	dueDateBlock.Optional = true

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			titleBlock,
			descriptionBlock,
			approverBlock,
			priorityBlock,
			categoryBlock,
			attachmentsBlock,
			dueDateBlock,
		},
	}

	privateMetadata := fmt.Sprintf("%s|%s", channelID, userID)

	modalRequest := slack.ModalViewRequest{
		Type:            slack.ViewType("modal"),
		Title:           slack.NewTextBlockObject(slack.PlainTextType, constants.ModalTitleCreateApproval, false, false),
		Close:           slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false),
		Submit:          slack.NewTextBlockObject(slack.PlainTextType, constants.ModalSubmitCreateApproval, false, false),
		Blocks:          blocks,
		CallbackID:      constants.ModalCallbackCreateApproval,
		PrivateMetadata: privateMetadata,
	}

	return modalRequest, nil
}

func (sh *SocketHandler) handleCreateApprovalSubmission(evt socketmode.Event, callback slack.InteractionCallback) {
	metadata := strings.Split(callback.View.PrivateMetadata, "|")
	if len(metadata) != 2 {
		log.Printf("Invalid private metadata: %s", callback.View.PrivateMetadata)
		sh.client.Ack(*evt.Request)
		return
	}

	channelID := metadata[0]
	requesterID := metadata[1]

	values := callback.View.State.Values

	title := strings.TrimSpace(values[constants.InputBlockTitle][constants.InputActionTitle].Value)
	description := strings.TrimSpace(values[constants.InputBlockDescription][constants.InputActionDescription].Value)
	approverID := values[constants.InputBlockApprover][constants.InputActionApprover].SelectedUser
	priority := values[constants.InputBlockPriority][constants.InputActionPriority].SelectedOption.Value
	category := values[constants.InputBlockCategory][constants.InputActionCategory].SelectedOption.Value

	attachments := ""
	if attachmentInput, ok := values[constants.InputBlockAttachments][constants.InputActionAttachments]; ok {
		attachments = strings.TrimSpace(attachmentInput.Value)
	}

	var dueDate *time.Time
	if dueDateInput, ok := values[constants.InputBlockDueDate][constants.InputActionDueDate]; ok && dueDateInput.SelectedDate != "" {
		parsedDate, err := time.Parse("2006-01-02", dueDateInput.SelectedDate)
		if err == nil {
			dueDate = &parsedDate
		}
	}

	if title == "" || description == "" || approverID == "" || priority == "" || category == "" {
		errors := make(map[string]string)
		if title == "" {
			errors[constants.InputBlockTitle] = constants.ErrorTitleRequired
		}
		if description == "" {
			errors[constants.InputBlockDescription] = constants.ErrorDescriptionRequired
		}
		if approverID == "" {
			errors[constants.InputBlockApprover] = constants.ErrorApproverRequired
		}
		if priority == "" {
			errors[constants.InputBlockPriority] = constants.ErrorPriorityRequired
		}
		if category == "" {
			errors[constants.InputBlockCategory] = constants.ErrorCategoryRequired
		}

		sh.client.Ack(*evt.Request, map[string]interface{}{
			"response_action": "errors",
			"errors":          errors,
		})
		return
	}

	err := sh.approvalService.CreateApprovalFromSlashCommand(
		channelID,
		requesterID,
		approverID,
		title,
		description,
		priority,
		category,
		attachments,
		dueDate,
	)

	if err != nil {
		log.Printf("Error creating approval from slash command: %v", err)
		sh.client.Ack(*evt.Request, map[string]interface{}{
			"response_action": "errors",
			"errors": map[string]string{
				constants.InputBlockTitle: "Error creating approval: " + err.Error(),
			},
		})
		return
	}

	sh.client.Ack(*evt.Request)
	log.Printf("Approval request created via slash command by user: %s", requesterID)
}
