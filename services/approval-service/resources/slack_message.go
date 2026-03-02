package resources

type Mention struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type SendMessageRequest struct {
	ChannelID   string    `json:"channel_id,omitempty"`
	ChannelName string    `json:"channel_name,omitempty"`
	Text        string    `json:"text"`
	Mentions    []Mention `json:"mentions,omitempty"`
	ThreadTS    string    `json:"thread_ts,omitempty"` //for threaded reply
}

type SendMessageResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ChannelID string `json:"channel_id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"` //for threaded msg
	Text      string `json:"text,omitempty"`
}
