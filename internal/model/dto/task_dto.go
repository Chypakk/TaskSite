package dto

import "time"

type TaskDTO struct {
	ID              int       `json:"id"`
	GroupID         *int      `json:"group_id,omitempty"`
	Username        string    `json:"username,omitempty"`
	GroupName       string    `json:"group_name,omitempty"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Author          string    `json:"author"`
	Status          string    `json:"status"`
	SolutionComment *string   `json:"solution_comment,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
	CompletedAt     time.Time `json:"completed_at,omitempty"`
}
