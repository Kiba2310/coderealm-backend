package types

import "time"

type GroupModel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TeacherID   string    `json:"teacher_id"`
	InviteCode  string    `json:"invite_code"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type JoinGroupRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

type APIResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
