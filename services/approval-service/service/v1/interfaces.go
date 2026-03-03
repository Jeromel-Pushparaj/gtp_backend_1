package v1

import (
	"github.com/jeromelp/gtp_backend_1/services/approval-service/resources"
	"github.com/slack-go/slack"
)

type SlackServiceInterface interface {
	CreateChannel(channelName string, isPrivate bool, description string) (string, error)
	AddMemberToChannel(channelID, userID string) error
	GetAllUsers() ([]resources.User, error)
	GetUserByName(userName string) (*resources.User, error)
	GetUserByID(userID string) (*resources.User, error)
	GetAllChannels() ([]resources.Channel, error)
	GetChannelByName(channelName string) (*resources.Channel, error)
	GetChannelByID(channelID string) (*resources.Channel, error)
	SendMessage(channelID, text string) (string, error)
	SendMessageWithMentions(channelID, text string, mentions []resources.Mention) (string, error)
	SendMessageInThread(channelID, text, threadTS string) (string, error)
	SendBlockMessage(channelID string, blocks []slack.Block, fallbackText string) (string, error)
	UpdateMessage(channelID, timestamp string, blocks []slack.Block) error
	SendApprovalFormButton(channelID string) (string, error)
}
