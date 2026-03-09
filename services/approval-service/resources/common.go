package resources

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
