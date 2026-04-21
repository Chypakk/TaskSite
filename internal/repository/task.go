package repository

import (
	"context"
	"tasksite/internal/model"
	"tasksite/internal/storage"
)

type TaskRepository interface {
	GetTasks(ctx context.Context, taskID, groupID *int, limit, offset int, statusFilter *string, sortBy string) ([]storage.TaskWithRelations, error)

	CreateTask(ctx context.Context, name, description, author string) (*model.Task, error)

	ClaimTask(ctx context.Context, taskId, userId int) error
	//CompleteTask(ctx context.Context, taskID, userID int) (*model.Task, error)
	CompleteTask(ctx context.Context, taskID, userID int) error
	DeleteTask(ctx context.Context, id, userId int) error
	UpdateTask(ctx context.Context, taskID int, req model.UpdateTaskRequest, editorID int) (*storage.TaskWithRelations, error)

	CountTasks(ctx context.Context, statusFilter *string, groupID *int) (int, error)
}