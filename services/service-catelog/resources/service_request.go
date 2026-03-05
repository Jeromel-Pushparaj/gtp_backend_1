package resources

// CreateServiceRequest represents a request to create/fetch a service
type CreateServiceRequest struct {
	OrgID int64 `form:"org_id" binding:"required"`
}

// GetServiceRequest represents a request to get a specific service by ID
type GetServiceRequest struct {
	ID string `uri:"id" binding:"required"`
}
