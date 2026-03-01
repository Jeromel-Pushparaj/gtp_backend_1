package v1

import (
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
