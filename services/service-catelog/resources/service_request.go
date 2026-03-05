package resources

// CreateServiceRequest represents a request to create/fetch a service
type CreateServiceRequest struct {
	OrgID int64 `uri:"org_id" binding:"required"`
}

// CreateServiceRequestLegacy represents a legacy request to create/fetch a service (query param)
type CreateServiceRequestLegacy struct {
	OrgID int64 `form:"org_id" binding:"required"`
}

// GetServiceRequest represents a request to get a specific service by ID
type GetServiceRequest struct {
	OrgID int64  `uri:"org_id" binding:"required"`
	ID    string `uri:"id" binding:"required"`
}

// GetServiceRequestLegacy represents a legacy request to get a specific service by ID
type GetServiceRequestLegacy struct {
	ID string `uri:"id" binding:"required"`
}

// GetAllServicesRequest represents a request to get all services for an organization
type GetAllServicesRequest struct {
	OrgID int64 `uri:"org_id" binding:"required"`
}
