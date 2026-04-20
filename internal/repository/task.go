package repository

import (
	"context"
	"tasksite/internal/model"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, name, description, author string) (*model.Task, error)
	GetTasks(ctx context.Context, statusFilter *string) ([]model.Task, error)
	GetTaskByID(ctx context.Context, taskID int) (*model.Task, error)
	ClaimTask(ctx context.Context, taskId, userId int) error
	CompleteTask(ctx context.Context, taskID, userID int) (*model.Task, error)
	DeleteTask(ctx context.Context, id, userId int) error
	UpdateTask(ctx context.Context, taskID int, req model.UpdateTaskRequest, editorID int) (*model.Task, error)
	GetUngroupedTasks(ctx context.Context, statusFilter *string) ([]model.Task, error)
	CountTasks(ctx context.Context, statusFilter *string, groupID *int) (int, error)
	GetTasksPaginated(ctx context.Context, pq model.PaginationQuery, groupID *int) ([]model.Task, error)
}