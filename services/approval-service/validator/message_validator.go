package validator

import (
	"fmt"

	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
)

func ValidateMessageText(text string) error {
	if text == "" {
		return fmt.Errorf(constants.ErrorMessageTextRequired)
	}
	if len(text) > 40000 {
		return fmt.Errorf("message text cannot exceed 40000 characters")
	}
	return nil
}

func ValidateMentionType(mentionType string) error {
	validTypes := map[string]bool{
		constants.MentionTypeUser:          true,
		constants.MentionTypeUserByName:    true,
		constants.MentionTypeChannelLink:   true,
		constants.MentionTypeChannelByName: true,
		constants.MentionTypeHere:          true,
		constants.MentionTypeChannelNotify: true,
		constants.MentionTypeEveryone:      true,
		constants.MentionTypeUserGroup:     true,
	}

	if !validTypes[mentionType] {
		return fmt.Errorf(constants.ErrorInvalidMentionType)
	}
	return nil
}
