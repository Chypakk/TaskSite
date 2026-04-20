package repository

import (
	"context"
	"tasksite/internal/model"
)

type TaskGroupRepository interface {
	CreateTaskGroup(ctx context.Context, name, description string) (*model.TaskGroup, error)
	GetTaskGroups(ctx context.Context) ([]model.TaskGroup, error)
	GetTaskGroupById(ctx context.Context, id int) (*model.TaskGroup, error)
	AssignTaskToGroup(ctx context.Context, taskID, groupID int) error
	RemoveTaskFromGroup(ctx context.Context, taskID int) error
	GetTasksByGroup(ctx context.Context, groupID int, statusFilter *string) ([]model.Task, error)
}