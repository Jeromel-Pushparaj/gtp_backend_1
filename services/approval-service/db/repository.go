package db

import (
	"time"

	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/models"
)

type ApprovalRepository struct {
	db *Database
}

func NewApprovalRepository(db *Database) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

func (r *ApprovalRepository) Create(approval *models.ApprovalRequest) error {
	return r.db.DB.Create(approval).Error
}

func (r *ApprovalRepository) GetByRequestID(requestID string) (*models.ApprovalRequest, error) {
	var approval models.ApprovalRequest
	err := r.db.DB.Where("request_id = ?", requestID).First(&approval).Error
	return &approval, err
}

func (r *ApprovalRepository) GetByID(id uint) (*models.ApprovalRequest, error) {
	var approval models.ApprovalRequest
	err := r.db.DB.First(&approval, id).Error
	return &approval, err
}

func (r *ApprovalRepository) UpdateStatus(requestID string, status string, approved bool, processedBy string, reason string) error {
	now := time.Now()
	return r.db.DB.Model(&models.ApprovalRequest{}).
		Where("request_id = ?", requestID).
		Updates(map[string]interface{}{
			"status":       status,
			"approved":     approved,
			"processed_by": processedBy,
			"processed_at": &now,
			"reason":       reason,
		}).Error
}

func (r *ApprovalRepository) GetPendingApprovals() ([]models.ApprovalRequest, error) {
	var approvals []models.ApprovalRequest
	err := r.db.DB.Where("status = ?", constants.ApprovalStatusPending).Find(&approvals).Error
	return approvals, err
}

func (r *ApprovalRepository) GetAll() ([]models.ApprovalRequest, error) {
	var approvals []models.ApprovalRequest
	err := r.db.DB.Find(&approvals).Error
	return approvals, err
}

func (r *ApprovalRepository) UpdateMessageTS(requestID string, messageTS string) error {
	return r.db.DB.Model(&models.ApprovalRequest{}).
		Where("request_id = ?", requestID).
		Update("message_ts", messageTS).Error
}
