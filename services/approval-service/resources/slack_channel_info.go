package resources

type Channel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
	IsMember  bool   `json:"is_member"`
}

type GetAllChannelsResponse struct {
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	Channels []Channel `json:"channels,omitempty"`
	Count    int       `json:"count"`
}

type GetChannelRequest struct {
	ChannelName string `json:"channel_name"`
	ChannelID   string `json:"channel_id"`
}

type GetChannelResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Channel *Channel `json:"channel,omitempty"`
}
