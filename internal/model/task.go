package model

import "time"

type Task struct {
	ID              int       `json:"id"`
	UserID          *int      `json:"user_id,omitempty"`
	GroupID         *int      `json:"group_id,omitempty"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Author          string    `json:"author"`
	Status          string    `json:"status"`
	SolutionComment *string   `json:"solution_comment,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
	CompletedAt     time.Time `json:"completed_at,omitempty"`
}

// CreateTaskRequest запрос создания задачи
type CreateTaskRequest struct {
	Name        string `json:"name" example:"Задача 1"`
	Description string `json:"description" example:"Описание"`
	Author      string `json:"author" example:"Пользователь1"`
}

type UpdateTaskRequest struct {
	Name            string `json:"name,omitempty" example:"Задача 1"`
	Description     string `json:"description,omitempty" example:"Описание"`
	Author          string `json:"author,omitempty" example:"Пользователь1"`
	Status          string `json:"status,omitempty" example:"open"`
	SolutionComment string `json:"solution_comment,omitempty" example:"что-то сделал"`
}
