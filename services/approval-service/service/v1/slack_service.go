package v1

import (
	"fmt"
	"log"
	"strings"

	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/resources"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/validator"
	"github.com/slack-go/slack"
)

type SlackService struct {
	client *slack.Client
}

func NewSlackService(token string) *SlackService {
	return &SlackService{
		client: slack.New(token),
	}
}

func (s *SlackService) CreateChannel(channelName string, isPrivate bool, description string) (string, error) {
	channel, err := s.client.CreateConversation(slack.CreateConversationParams{
		ChannelName: channelName,
		IsPrivate:   isPrivate,
	})

	if err != nil {
		return "", fmt.Errorf("%s: %w", constants.ErrorChannelCreationFailed, err)
	}

	if description != "" {
		_, err = s.client.SetTopicOfConversation(channel.ID, description)
		if err != nil {
			log.Printf("Warning: %s: %v\n", constants.ErrorTopicSetFailed, err)
		}
	}

	return channel.ID, nil
}

func (s *SlackService) AddMemberToChannel(channelID, userID string) error {
	_, err := s.client.InviteUsersToConversation(channelID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", constants.ErrorAddMemberFailed, err)
	}
	return nil
}

func (s *SlackService) GetAllUsers() ([]resources.User, error) {
	users, err := s.client.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", constants.ErrorGetUsersFailed, err)
	}

	var userList []resources.User
	for _, user := range users {
		if !user.IsBot && !user.Deleted {
			userList = append(userList, resources.User{
				ID:       user.ID,
				Name:     user.Name,
				RealName: user.RealName,
				Email:    user.Profile.Email,
			})
		}
	}

	return userList, nil
}

func (s *SlackService) GetUserByName(userName string) (*resources.User, error) {

	users, err := s.GetAllUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if strings.EqualFold(user.Name, userName) || strings.EqualFold(user.RealName, userName) {
			return &user, nil
		}
	}

	return nil, fmt.Errorf(constants.ErrorUserNotFound)
}

func (s *SlackService) GetUserByID(userID string) (*resources.User, error) {
	user, err := s.client.GetUserInfo(userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", constants.ErrorUserNotFound, err)
	}

	return &resources.User{
		ID:       user.ID,
		Name:     user.Name,
		RealName: user.RealName,
		Email:    user.Profile.Email,
	}, nil
}

func (s *SlackService) GetAllChannels() ([]resources.Channel, error) {
	params := &slack.GetConversationsParameters{
		ExcludeArchived: true,
		Types:           []string{"public_channel", "private_channel"},
	}

	channels, _, err := s.client.GetConversations(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", constants.ErrorGetChannelsFailed, err)
	}

	var channelList []resources.Channel
	for _, channel := range channels {
		channelList = append(channelList, resources.Channel{
			ID:        channel.ID,
			Name:      channel.Name,
			IsPrivate: channel.IsPrivate,
			IsMember:  channel.IsMember,
		})
	}

	return channelList, nil
}

func (s *SlackService) GetChannelByName(channelName string) (*resources.Channel, error) {
	channels, err := s.GetAllChannels()
	if err != nil {
		return nil, err
	}

	for _, channel := range channels {
		if strings.EqualFold(channel.Name, channelName) {
			return &channel, nil
		}
	}

	return nil, fmt.Errorf(constants.ErrorChannelNotFound)
}

func (s *SlackService) GetChannelByID(channelID string) (*resources.Channel, error) {
	channel, err := s.client.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", constants.ErrorChannelNotFound, err)
	}

	return &resources.Channel{
		ID:        channel.ID,
		Name:      channel.Name,
		IsPrivate: channel.IsPrivate,
		IsMember:  channel.IsMember,
	}, nil
}

func (s *SlackService) buildMentionString(mention resources.Mention) (string, error) {
	if err := validator.ValidateMentionType(mention.Type); err != nil {
		return "", err
	}

	switch mention.Type {
	case constants.MentionTypeUser:
		if mention.ID == "" {
			return "", fmt.Errorf("user ID is required for user mention")
		}
		return fmt.Sprintf("<@%s>", mention.ID), nil

	case constants.MentionTypeUserByName:
		if mention.Name == "" {
			return "", fmt.Errorf("user name is required for user mention")
		}
		user, err := s.GetUserByName(mention.Name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("<@%s>", user.ID), nil

	case constants.MentionTypeChannelLink:
		if mention.ID == "" {
			return "", fmt.Errorf("channel ID is required for channel link mention")
		}
		return fmt.Sprintf("<#%s>", mention.ID), nil

	case constants.MentionTypeChannelByName:
		if mention.Name == "" {
			return "", fmt.Errorf("channel name is required for channel_by_name mention")
		}
		channel, err := s.GetChannelByName(mention.Name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("<#%s|%s>", channel.ID, channel.Name), nil

	case constants.MentionTypeHere:
		return "<!here>", nil

	case constants.MentionTypeChannelNotify:
		return "<!channel>", nil

	case constants.MentionTypeEveryone:
		return "<!everyone>", nil

	case constants.MentionTypeUserGroup:
		if mention.ID == "" {
			return "", fmt.Errorf("user group ID is required for usergroup mention")
		}
		return fmt.Sprintf("<!subteam^%s>", mention.ID), nil

	default:
		return "", fmt.Errorf("invalid mention type: %s", mention.Type)
	}
}

func (s *SlackService) SendMessage(channelID, text string) (string, error) {
	_, timestamp, err := s.client.PostMessage(channelID, slack.MsgOptionText(text, false))
	if err != nil {
		return "", fmt.Errorf("%s: %w", constants.ErrorMessageSendFailed, err)
	}
	return timestamp, nil
}

func (s *SlackService) SendMessageWithMentions(channelID, text string, mentions []resources.Mention) (string, error) {
	var mentionStrings []string
	for _, mention := range mentions {
		mentionString, err := s.buildMentionString(mention)
		if err != nil {
			return "", err
		}
		mentionStrings = append(mentionStrings, mentionString)
	}

	finalText := text
	if len(mentionStrings) > 0 {
		var builder strings.Builder
		for _, m := range mentionStrings {
			builder.WriteString(m)
			builder.WriteString(" ")
		}
		builder.WriteString(text)
		finalText = builder.String()
	}

	_, timestamp, err := s.client.PostMessage(channelID, slack.MsgOptionText(finalText, false))
	if err != nil {
		return "", fmt.Errorf("%s: %w", constants.ErrorMessageSendFailed, err)
	}
	return timestamp, nil
}

func (s *SlackService) SendMessageInThread(channelID, text, threadTS string) (string, error) {
	_, timestamp, err := s.client.PostMessage(
		channelID,
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(threadTS),
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", constants.ErrorMessageSendFailed, err)
	}
	return timestamp, nil
}

func (s *SlackService) SendBlockMessage(channelID string, blocks []slack.Block, fallbackText string) (string, error) {
	_, timestamp, err := s.client.PostMessage(
		channelID,
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionText(fallbackText, false),
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", constants.ErrorMessageSendFailed, err)
	}
	return timestamp, nil
}

func (s *SlackService) UpdateMessage(channelID, timestamp string, blocks []slack.Block) error {
	_, _, _, err := s.client.UpdateMessage(
		channelID,
		timestamp,
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionText("", false),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", constants.ErrorMessageSendFailed, err)
	}
	return nil
}
