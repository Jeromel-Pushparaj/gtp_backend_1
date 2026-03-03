package validator

import (
	"fmt"
	"regexp"
)

var channelNameRegex = regexp.MustCompile(`^[a-z0-9-_]{1,80}$`)

func ValidateChannelName(name string) error {
	if name == "" {
		return fmt.Errorf("channel name cannot be empty")
	}
	if !channelNameRegex.MatchString(name) {
		return fmt.Errorf("channel name must be lowercase alphanumeric with dashes/underscores, max 80 chars")
	}
	return nil
}

func ValidateUserID(id string) error {
	if id == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	return nil
}

func ValidateChannelID(id string) error {
	if id == "" {
		return fmt.Errorf("channel ID cannot be empty")
	}
	return nil
}
