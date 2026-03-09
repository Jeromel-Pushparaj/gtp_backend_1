package resources

type CreateChannelRequest struct {
	ChannelName string `json:"channel_name"`
	IsPrivate   bool   `json:"is_private"`
	Description string `json:"description"`
}

type CreateChannelResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ChannelID   string `json:"channel_id,omitempty"`
	ChannelName string `json:"channel_name,omitempty"`
}
