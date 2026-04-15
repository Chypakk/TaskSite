package service

import (
	"context"
	"tasksite/internal/model"
	"tasksite/internal/model/dto"
	"tasksite/internal/storage"
)

type GroupService struct {
	storage *storage.Storage
}

func NewGroupService(storage *storage.Storage) *GroupService {
	return &GroupService{storage: storage}
}

func (g *GroupService) CreateTaskGroup(ctx context.Context, name, description string) (dto.TaskGroupDTO, error) {
	taskGroup, err := g.storage.CreateTaskGroup(ctx, name, description)
	if err != nil {
		return dto.TaskGroupDTO{}, err
	}

	dto := g.toDTO(ctx, *taskGroup)

	return dto, nil
}

func (g *GroupService) GetTaskGroups(ctx context.Context) ([]dto.TaskGroupDTO, error) {
	taskGroups, err := g.storage.GetTaskGroups(ctx)

	if err != nil {
		return []dto.TaskGroupDTO{}, err
	}

	taskGroupsDTO := make([]dto.TaskGroupDTO, len(taskGroups))
	for i, task := range taskGroups {
		taskGroupsDTO[i] = g.toDTO(ctx, task)
	}

	return taskGroupsDTO, nil
}

func (g *GroupService) AssignTaskToGroup(ctx context.Context, taskID, groupID int) error {
	if err := g.storage.AssignTaskToGroup(ctx, taskID, groupID); err != nil {
		return err
	}

	return nil
}

func (g *GroupService) RemoveTaskFromGroup(ctx context.Context, taskID int) error {
	if err := g.storage.RemoveTaskFromGroup(ctx, taskID); err != nil {
		return err
	}

	return nil
}

func (g *GroupService) GetTasksByGroup(ctx context.Context, groupID int, statusFilter *string) ([]dto.TaskDTO, error) {
	tasks, err := g.storage.GetTasksByGroup(ctx, groupID, statusFilter)
	if err != nil {
		return []dto.TaskDTO{}, err
	}

	tasksDTO := make([]dto.TaskDTO, len(tasks))
	for i, task := range tasks {
		tasksDTO[i] = g.taskToDTO(ctx, task)
	}

	return tasksDTO, nil
}

func (g *GroupService) toDTO(ctx context.Context, taskGroup model.TaskGroup) dto.TaskGroupDTO {
	var taskGroupDTO dto.TaskGroupDTO

	taskGroupDTO.ID = taskGroup.ID
	taskGroupDTO.Name = taskGroup.Name
	taskGroupDTO.Description = taskGroup.Description
	taskGroupDTO.CreatedAt = taskGroup.CreatedAt

	return taskGroupDTO
}

func (g *GroupService) taskToDTO(ctx context.Context, task model.Task) dto.TaskDTO {
	var taskDTO dto.TaskDTO
	username := ""
	if task.UserID != nil {
		user, err := g.storage.GetUserById(ctx, *task.UserID)
		if err == nil {
			username = user.Username
		}

	}

	taskDTO.ID = task.ID
	taskDTO.Name = task.Name
	taskDTO.Author = task.Author
	taskDTO.CompletedAt = task.CompletedAt
	taskDTO.CreatedAt = task.CreatedAt
	taskDTO.Description = task.Description
	taskDTO.Status = task.Status
	taskDTO.UpdatedAt = task.UpdatedAt
	taskDTO.Username = username

	return taskDTO
}