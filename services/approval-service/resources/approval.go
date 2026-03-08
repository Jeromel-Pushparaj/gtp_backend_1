package resources

import "time"

type ApprovalRequestMessage struct {
	RequestID     string                 `json:"request_id"`
	RequesterID   string                 `json:"requester_id"`
	RequesterName string                 `json:"requester_name"`
	ApproverID    string                 `json:"approver_id"`
	ApproverName  string                 `json:"approver_name"`
	ChannelID     string                 `json:"channel_id"`
	RequestType   string                 `json:"request_type"`
	RequestData   map[string]interface{} `json:"request_data"`
	Message       string                 `json:"message"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Priority      string                 `json:"priority"`
	Category      string                 `json:"category"`
	Attachments   string                 `json:"attachments"`
	DueDate       *time.Time             `json:"due_date"`
}

type CreateApprovalRequest struct {
	ChannelName   string                 `json:"channel_name" binding:"required"`
	ApproverName  string                 `json:"approver_name" binding:"required"`
	RequesterName string                 `json:"requester_name" binding:"required"`
	RequestType   string                 `json:"request_type" binding:"required"`
	Message       string                 `json:"message" binding:"required"`
	RequestData   map[string]interface{} `json:"request_data"`
}

type CreateApprovalResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type ApprovalCompletedMessage struct {
	RequestID       string                 `json:"request_id"`
	Status          string                 `json:"status"`
	Approved        bool                   `json:"approved"`
	ProcessedBy     string                 `json:"processed_by"`
	ProcessedAt     time.Time              `json:"processed_at"`
	Reason          string                 `json:"reason"`
	ApproverComment string                 `json:"approver_comment"`
	RequestData     map[string]interface{} `json:"request_data"`
}

type ActionMessage struct {
	RequestID   string                 `json:"request_id"`
	Action      string                 `json:"action"`
	Status      string                 `json:"status"`
	ProcessedBy string                 `json:"processed_by"`
	ProcessedAt time.Time              `json:"processed_at"`
	RequestData map[string]interface{} `json:"request_data"`
	Message     string                 `json:"message"`
	Reason      string                 `json:"reason,omitempty"`
}

type DomainChangeApprovalRequest struct {
	ApproverID     string `json:"approver_id"`
	ApproverName   string `json:"approver_name"`
	RequesterID    string `json:"requester_id"`
	RequesterName  string `json:"requester_name"`
	OldDomainName  string `json:"old_domain_name" binding:"required"`
	NewDomainName  string `json:"new_domain_name" binding:"required"`
	ChangeReason   string `json:"change_reason" binding:"required"`
	AdditionalInfo string `json:"additional_info"`
	ChannelID      string `json:"channel_id"`
	ChannelName    string `json:"channel_name"`
	UseAppDM       bool   `json:"use_app_dm"`
	AppBotUserID   string `json:"app_bot_user_id"`
}

type GetAppsResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Apps    []AppInfo `json:"apps"`
	Count   int       `json:"count"`
}

type AppInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsBot    bool   `json:"is_bot"`
	RealName string `json:"real_name"`
}

type GetDMChannelRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type GetDMChannelResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ChannelID string `json:"channel_id"`
}
