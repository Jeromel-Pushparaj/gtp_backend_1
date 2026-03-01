package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	controllerv1 "github.com/jeromelp/gtp_backend_1/services/approval-service/controller/v1"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/middleware"
)

func SetupRoutes(slackController *controllerv1.SlackController, approvalController *controllerv1.ApprovalController) *gin.Engine {
	router := gin.New()

	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())

	router.GET(constants.RouteHealth, healthCheckHandler)

	v1 := router.Group(constants.RouteAPIV1Prefix)
	{
		v1.POST(constants.RouteSlackChannelCreate, slackController.CreateChannel)
		v1.POST(constants.RouteSlackAddMember, slackController.AddMember)

		v1.GET(constants.RouteSlackGetAllUsers, slackController.GetAllUsers)
		v1.POST(constants.RouteSlackGetUserByName, slackController.GetUserByName)
		v1.POST(constants.RouteSlackGetUserByID, slackController.GetUserByID)

		v1.GET(constants.RouteSlackGetAllChannels, slackController.GetAllChannels)
		v1.POST(constants.RouteSlackGetChannelByName, slackController.GetChannelByName)
		v1.POST(constants.RouteSlackGetChannelByID, slackController.GetChannelByID)
		v1.POST(constants.RouteSlackSendMessage, slackController.SendMessage)

		v1.GET(constants.RouteApprovalGetAll, approvalController.GetAllApprovals)
		v1.GET(constants.RouteApprovalGetPending, approvalController.GetPendingApprovals)
		v1.POST(constants.RouteApprovalGetByID, approvalController.GetApprovalByID)
		v1.POST(constants.RouteApprovalCreate, approvalController.CreateApprovalRequest)
	}

	return router
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		constants.HealthStatusKey:  constants.HealthStatusValue,
		constants.HealthServiceKey: constants.HealthServiceValue,
	})
}
