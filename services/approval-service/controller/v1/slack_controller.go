package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/resources"
	v1 "github.com/jeromelp/gtp_backend_1/services/approval-service/service/v1"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/validator"
)

type SlackController struct {
	slackService v1.SlackServiceInterface
}

func NewSlackController(slackService v1.SlackServiceInterface) *SlackController {
	return &SlackController{
		slackService: slackService,
	}
}

func (sc *SlackController) CreateChannel(c *gin.Context) {
	var req resources.CreateChannelRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody,
		})
		return
	}

	if err := validator.ValidateChannelName(req.ChannelName); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	channelID, err := sc.slackService.CreateChannel(req.ChannelName, req.IsPrivate, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, resources.CreateChannelResponse{
		Success:     true,
		Message:     constants.SuccessChannelCreated,
		ChannelID:   channelID,
		ChannelName: req.ChannelName,
	})
}

func (sc *SlackController) AddMember(c *gin.Context) {
	var req resources.AddMemberRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody,
		})
		return
	}

	var channelID, userID string
	var err error
	var channel *resources.Channel
	var user *resources.User

	if req.ChannelID != "" {
		if err := validator.ValidateChannelID(req.ChannelID); err != nil {
			c.JSON(http.StatusBadRequest, resources.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		channelID = req.ChannelID
	} else if req.ChannelName != "" {
		channel, err = sc.slackService.GetChannelByName(req.ChannelName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
				Success: false,
				Error:   constants.ErrorChannelNotFound,
			})
			return
		}
		channelID = channel.ID
	} else {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorChannelIDRequired,
		})
		return
	}

	if req.UserID != "" {
		if err := validator.ValidateUserID(req.UserID); err != nil {
			c.JSON(http.StatusBadRequest, resources.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		userID = req.UserID
	} else if req.UserName != "" {
		user, err = sc.slackService.GetUserByName(req.UserName)
		if err != nil {
			c.JSON(http.StatusNotFound, resources.ErrorResponse{
				Success: false,
				Error:   constants.ErrorUserNotFound,
			})
			return
		}
		userID = user.ID
	} else {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorUserIDRequired,
		})
		return
	}

	err = sc.slackService.AddMemberToChannel(channelID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, resources.AddMemberResponse{
		Success:     true,
		Message:     constants.SuccessMemberAdded,
		ChannelID:   channelID,
		ChannelName: req.ChannelName,
		UserID:      userID,
		UserName:    req.UserName,
	})
}

func (sc *SlackController) GetAllUsers(c *gin.Context) {
	users, err := sc.slackService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resources.GetAllUsersResponse{
		Success: true,
		Message: constants.SuccessUsersFetched,
		Users:   users,
		Count:   len(users),
	})
}

func (sc *SlackController) GetUserByName(c *gin.Context) {
	var req resources.GetUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody,
		})
		return
	}
	if req.UserName == "" {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorUserNameRequired,
		})
		return
	}
	user, err := sc.slackService.GetUserByName(req.UserName)
	if err != nil {
		c.JSON(http.StatusNotFound, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorUserNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, resources.GetUserResponse{
		Success: true,
		Message: constants.SuccessUserFetched,
		User:    user,
	})
}

func (sc *SlackController) GetUserByID(c *gin.Context) {
	var req resources.GetUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody,
		})
		return
	}

	if err := validator.ValidateUserID(req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	user, err := sc.slackService.GetUserByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorUserNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, resources.GetUserResponse{
		Success: true,
		Message: constants.SuccessUserFetched,
		User:    user,
	})
}

func (sc *SlackController) GetAllChannels(c *gin.Context) {
	channels, err := sc.slackService.GetAllChannels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resources.GetAllChannelsResponse{
		Success:  true,
		Message:  constants.SuccessChannelsFetched,
		Channels: channels,
		Count:    len(channels),
	})
}

func (sc *SlackController) GetChannelByName(c *gin.Context) {
	var req resources.GetChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody,
		})
		return
	}
	if req.ChannelName == "" {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorChannelNameRequired,
		})
		return
	}
	channel, err := sc.slackService.GetChannelByName(req.ChannelName)
	if err != nil {
		c.JSON(http.StatusNotFound, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorChannelNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, resources.GetChannelResponse{
		Success: true,
		Message: constants.SuccessChannelFetched,
		Channel: channel,
	})
}

func (sc *SlackController) GetChannelByID(c *gin.Context) {
	var req resources.GetChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody,
		})
		return
	}

	if err := validator.ValidateChannelID(req.ChannelID); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	channel, err := sc.slackService.GetChannelByID(req.ChannelID)
	if err != nil {
		c.JSON(http.StatusNotFound, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorChannelNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, resources.GetChannelResponse{
		Success: true,
		Message: constants.SuccessChannelFetched,
		Channel: channel,
	})
}

func (sc *SlackController) SendMessage(c *gin.Context) {
	var req resources.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody,
		})
		return
	}

	if err := validator.ValidateMessageText(req.Text); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	var channelID string
	var err error
	if req.ChannelID != "" {
		if err := validator.ValidateChannelID(req.ChannelID); err != nil {
			c.JSON(http.StatusBadRequest, resources.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		channelID = req.ChannelID
	} else if req.ChannelName != "" {
		channel, err := sc.slackService.GetChannelByName(req.ChannelName)
		if err != nil {
			c.JSON(http.StatusNotFound, resources.ErrorResponse{
				Success: false,
				Error:   constants.ErrorChannelNotFound,
			})
			return
		}
		channelID = channel.ID
	} else {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorChannelIDRequired,
		})
		return
	}
	var timestamp string
	if req.ThreadTS != "" {
		timestamp, err = sc.slackService.SendMessageInThread(channelID, req.Text, req.ThreadTS)
	} else if len(req.Mentions) > 0 {
		timestamp, err = sc.slackService.SendMessageWithMentions(channelID, req.Text, req.Mentions)
	} else {
		timestamp, err = sc.slackService.SendMessage(channelID, req.Text)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resources.SendMessageResponse{
		Success:   true,
		Message:   constants.SuccessMessageSent,
		ChannelID: channelID,
		Timestamp: timestamp,
		Text:      req.Text,
	})
}

func (sc *SlackController) SendApprovalFormButton(c *gin.Context) {
	var req resources.GetChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody,
		})
		return
	}

	var channelID string
	if req.ChannelID != "" {
		if err := validator.ValidateChannelID(req.ChannelID); err != nil {
			c.JSON(http.StatusBadRequest, resources.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		channelID = req.ChannelID
	} else if req.ChannelName != "" {
		channel, err := sc.slackService.GetChannelByName(req.ChannelName)
		if err != nil {
			c.JSON(http.StatusNotFound, resources.ErrorResponse{
				Success: false,
				Error:   constants.ErrorChannelNotFound,
			})
			return
		}
		channelID = channel.ID
	} else {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorChannelIDRequired,
		})
		return
	}

	timestamp, err := sc.slackService.SendApprovalFormButton(channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resources.SendMessageResponse{
		Success:   true,
		Message:   "Approval form button sent successfully",
		ChannelID: channelID,
		Timestamp: timestamp,
		Text:      "Click to create a new approval request",
	})
}
