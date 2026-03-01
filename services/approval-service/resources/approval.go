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
