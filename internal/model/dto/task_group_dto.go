package dto

import "time"

type TaskGroupDTO struct {
	ID          int       `json:"group_id"`
	Name        string    `json:"group_name"`
	Description string    `json:"group_desc,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}