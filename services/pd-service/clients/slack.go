package clients

import (
	"context"
	"fmt"
	"pd-service/models"

	"github.com/slack-go/slack"
)

type SlackClient struct {
	client *slack.Client
}

func NewSlackClient(token string) *SlackClient {
	return &SlackClient{
		client: slack.New(token),
	}
}

func (s *SlackClient) ListUsers(ctx context.Context) ([]*models.SlackUser, error) {
	users, err := s.client.GetUsersContext(ctx)
	if err != nil {
		return nil, err
	}
	
	var result []*models.SlackUser
	for _, u := range users {
		if u.IsBot || u.Deleted {
			continue
		}
		
		result = append(result, &models.SlackUser{
			ID:       u.ID,
			Name:     u.Name,
			RealName: u.RealName,
			Email:    u.Profile.Email,
		})
	}
	
	return result, nil
}

func (s *SlackClient) SendMessage(ctx context.Context, userID, message string) error {
	conv, _, _, err := s.client.OpenConversationContext(ctx, &slack.OpenConversationParameters{
		Users: []string{userID},
	})
	if err != nil {
		return err
	}

	channelID := conv.ID
	if channelID == "" {
		channelID = userID
	}

	_, _, err = s.client.PostMessageContext(ctx, channelID, slack.MsgOptionText(message, false))
	return err
}

func (s *SlackClient) SendIncidentNotification(ctx context.Context, userID, serviceName, incidentTitle, incidentURL string) error {
	message := fmt.Sprintf(
		"🚨 *Test Incident Triggered*\n\n"+
			"*Service:* %s\n"+
			"*Incident:* %s\n"+
			"*URL:* %s\n\n"+
			"This is a test incident from the PD Service Dashboard.",
		serviceName, incidentTitle, incidentURL,
	)
	
	return s.SendMessage(ctx, userID, message)
}

