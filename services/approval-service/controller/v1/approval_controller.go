package v1

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/db"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/resources"
	servicev1 "github.com/jeromelp/gtp_backend_1/services/approval-service/service/v1"
)

type ApprovalController struct {
	repo         *db.ApprovalRepository
	slackService servicev1.SlackServiceInterface
	kafkaService *servicev1.KafkaService
}

func NewApprovalController(repo *db.ApprovalRepository, slackService servicev1.SlackServiceInterface, kafkaService *servicev1.KafkaService) *ApprovalController {
	return &ApprovalController{
		repo:         repo,
		slackService: slackService,
		kafkaService: kafkaService,
	}
}

func (ac *ApprovalController) GetAllApprovals(c *gin.Context) {
	approvals, err := ac.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch approvals",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Approvals fetched successfully",
		"approvals": approvals,
		"count":     len(approvals),
	})
}

func (ac *ApprovalController) GetPendingApprovals(c *gin.Context) {
	approvals, err := ac.repo.GetPendingApprovals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch pending approvals",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Pending approvals fetched successfully",
		"approvals": approvals,
		"count":     len(approvals),
	})
}

func (ac *ApprovalController) GetApprovalByID(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody,
		})
		return
	}

	approval, err := ac.repo.GetByID(req.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorApprovalNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Approval fetched successfully",
		"approval": approval,
	})
}

func (ac *ApprovalController) CreateApprovalRequest(c *gin.Context) {
	var req resources.CreateApprovalRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody,
		})
		return
	}

	channel, err := ac.slackService.GetChannelByName(req.ChannelName)
	if err != nil {
		c.JSON(http.StatusNotFound, resources.ErrorResponse{
			Success: false,
			Error:   "Channel not found: " + req.ChannelName,
		})
		return
	}

	approver, err := ac.slackService.GetUserByName(req.ApproverName)
	if err != nil {
		c.JSON(http.StatusNotFound, resources.ErrorResponse{
			Success: false,
			Error:   "Approver not found: " + req.ApproverName,
		})
		return
	}

	requester, err := ac.slackService.GetUserByName(req.RequesterName)
	if err != nil {
		c.JSON(http.StatusNotFound, resources.ErrorResponse{
			Success: false,
			Error:   "Requester not found: " + req.RequesterName,
		})
		return
	}

	requestID := uuid.New().String()

	approvalMsg := resources.ApprovalRequestMessage{
		RequestID:     requestID,
		RequesterID:   requester.ID,
		RequesterName: requester.RealName,
		ApproverID:    approver.ID,
		ApproverName:  approver.RealName,
		ChannelID:     channel.ID,
		RequestType:   req.RequestType,
		RequestData:   req.RequestData,
		Message:       req.Message,
	}

	err = ac.kafkaService.Publish(constants.KafkaTopicApprovalRequested, requestID, approvalMsg)
	if err != nil {
		log.Printf("Error publishing to Kafka: %v", err)
		c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
			Success: false,
			Error:   "Failed to publish approval request",
		})
		return
	}

	log.Printf("Published approval request to Kafka: %s", requestID)

	c.JSON(http.StatusOK, resources.CreateApprovalResponse{
		Success:   true,
		Message:   "Approval request created and published to Kafka",
		RequestID: requestID,
	})
}

func (ac *ApprovalController) CreateDomainChangeApprovalRequest(c *gin.Context) {
	var req resources.DomainChangeApprovalRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   constants.ErrorInvalidRequestBody + ": " + err.Error(),
		})
		return
	}

	var approver *resources.User
	var requester *resources.User
	var err error

	if req.ApproverID != "" {
		approver, err = ac.slackService.GetUserByID(req.ApproverID)
		if err != nil {
			c.JSON(http.StatusNotFound, resources.ErrorResponse{
				Success: false,
				Error:   "Approver not found: " + req.ApproverID,
			})
			return
		}
	} else if req.ApproverName != "" {
		approver, err = ac.slackService.GetUserByName(req.ApproverName)
		if err != nil {
			c.JSON(http.StatusNotFound, resources.ErrorResponse{
				Success: false,
				Error:   "Approver not found: " + req.ApproverName,
			})
			return
		}

		req.ApproverID = approver.ID
	} else {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   "Either approver_id or approver_name is required",
		})
		return
	}

	if req.RequesterID != "" {
		requester, err = ac.slackService.GetUserByID(req.RequesterID)
		if err != nil {
			c.JSON(http.StatusNotFound, resources.ErrorResponse{
				Success: false,
				Error:   "Requester not found: " + req.RequesterID,
			})
			return
		}
	} else if req.RequesterName != "" {
		requester, err = ac.slackService.GetUserByName(req.RequesterName)
		if err != nil {
			c.JSON(http.StatusNotFound, resources.ErrorResponse{
				Success: false,
				Error:   "Requester not found: " + req.RequesterName,
			})
			return
		}

		req.RequesterID = requester.ID
	} else {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   "Either requester_id or requester_name is required",
		})
		return
	}

	log.Printf("Resolved approver: %s (ID: %s)", approver.RealName, req.ApproverID)
	log.Printf("Resolved requester: %s (ID: %s)", requester.RealName, req.RequesterID)

	requestID := uuid.New().String()

	message := fmt.Sprintf("*Domain Name Change Request*\n\n"+
		"*Old Domain:* `%s`\n"+
		"*New Domain:* `%s`\n\n"+
		"*Reason:* %s",
		req.OldDomainName,
		req.NewDomainName,
		req.ChangeReason,
	)

	if req.AdditionalInfo != "" {
		message += fmt.Sprintf("\n\n*Additional Information:*\n%s", req.AdditionalInfo)
	}

	requestData := map[string]interface{}{
		"old_domain_name": req.OldDomainName,
		"new_domain_name": req.NewDomainName,
		"change_reason":   req.ChangeReason,
		"additional_info": req.AdditionalInfo,
		"change_type":     "domain_name_change",
		"is_app_dm":       req.UseAppDM,
	}

	var channelID string
	if req.UseAppDM {

		if req.AppBotUserID == "" {
			c.JSON(http.StatusBadRequest, resources.ErrorResponse{
				Success: false,
				Error:   "app_bot_user_id is required when use_app_dm is true",
			})
			return
		}

		log.Printf("Attempting to open DM channel with approver: %s (ID: %s)", approver.RealName, req.ApproverID)
		dmChannelID, err := ac.slackService.OpenDMChannel(req.ApproverID)
		if err != nil {
			log.Printf("ERROR: Failed to open DM channel with %s (%s): %v", approver.RealName, req.ApproverID, err)
			c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
				Success: false,
				Error:   "Failed to open DM channel: " + err.Error(),
			})
			return
		}
		channelID = dmChannelID
		log.Printf("Successfully opened DM channel with approver %s (%s): %s", approver.RealName, req.ApproverID, dmChannelID)
	} else if req.ChannelID != "" {

		channelID = req.ChannelID
	} else if req.ChannelName != "" {

		channel, err := ac.slackService.GetChannelByName(req.ChannelName)
		if err != nil {
			c.JSON(http.StatusNotFound, resources.ErrorResponse{
				Success: false,
				Error:   "Channel not found: " + req.ChannelName,
			})
			return
		}
		channelID = channel.ID
	} else {
		c.JSON(http.StatusBadRequest, resources.ErrorResponse{
			Success: false,
			Error:   "Either channel_id, channel_name, or use_app_dm must be provided",
		})
		return
	}

	approvalMsg := resources.ApprovalRequestMessage{
		RequestID:     requestID,
		RequesterID:   requester.ID,
		RequesterName: requester.RealName,
		ApproverID:    approver.ID,
		ApproverName:  approver.RealName,
		ChannelID:     channelID,
		RequestType:   "domain_change",
		RequestData:   requestData,
		Message:       message,
		Title:         "Domain Name Change Request",
		Description:   fmt.Sprintf("Change domain from %s to %s", req.OldDomainName, req.NewDomainName),
		Priority:      "high",
		Category:      "infrastructure",
	}

	err = ac.kafkaService.Publish(constants.KafkaTopicApprovalRequested, requestID, approvalMsg)
	if err != nil {
		log.Printf("Error publishing to Kafka: %v", err)
		c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
			Success: false,
			Error:   "Failed to publish approval request",
		})
		return
	}

	log.Printf("Published domain change approval request to Kafka: %s (Old: %s, New: %s, Approver: %s, Channel: %s)",
		requestID, req.OldDomainName, req.NewDomainName, approver.RealName, channelID)

	c.JSON(http.StatusOK, resources.CreateApprovalResponse{
		Success:   true,
		Message:   "Domain change approval request created and published to Kafka",
		RequestID: requestID,
	})
}
