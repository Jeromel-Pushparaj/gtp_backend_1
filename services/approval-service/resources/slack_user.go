package resources

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
	Email    string `json:"email"`
}

type GetAllUsersResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Users   []User `json:"users,omitempty"`
	Count   int    `json:"count"`
}

type GetUserRequest struct {
	UserName string `json:"user_name"`
	UserID   string `json:"user_id"`
}

type GetUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	User    *User  `json:"user,omitempty"`
}
