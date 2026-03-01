package models

import (
	"time"

	"gorm.io/gorm"
)

type ApprovalRequest struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	RequestID     string         `gorm:"uniqueIndex;not null" json:"request_id"`
	RequesterID   string         `gorm:"not null" json:"requester_id"`
	RequesterName string         `json:"requester_name"`
	ApproverID    string         `gorm:"not null" json:"approver_id"`
	ApproverName  string         `json:"approver_name"`
	ChannelID     string         `gorm:"not null" json:"channel_id"`
	MessageTS     string         `json:"message_ts"`
	RequestType   string         `json:"request_type"`
	RequestData   string         `gorm:"type:text" json:"request_data"`
	Status        string         `gorm:"default:'pending'" json:"status"`
	Approved      *bool          `json:"approved"`
	ProcessedBy   string         `json:"processed_by"`
	ProcessedAt   *time.Time     `json:"processed_at"`
	Reason        string         `json:"reason"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ApprovalRequest) TableName() string {
	return "approval_requests"
}
