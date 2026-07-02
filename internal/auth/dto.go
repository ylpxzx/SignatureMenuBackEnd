package auth

import "time"

type credentialsRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type profileUpdateRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type userResponse struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}
