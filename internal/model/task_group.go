package model

import "time"

type TaskGroup struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Для создания группы
type CreateGroupRequest struct {
    Name        string `json:"name" example:"ИИ ТП"`
    Description string `json:"description,omitempty" example:"Задачи по ИИ"`
}

// Для привязки задачи к группе
type AssignTaskToGroupRequest struct {
    GroupID int `json:"group_id" example:"1"`
}