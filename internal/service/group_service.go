package service

import (
	"context"
	"tasksite/internal/model"
	"tasksite/internal/model/dto"
	"tasksite/internal/repository"
	"tasksite/internal/storage"
)

type GroupService struct {
	repo repository.TaskGroupRepository
	userRepo repository.UserRepository
}

func NewGroupService(storage *storage.Storage) *GroupService {
	return &GroupService{
		repo: storage,
		userRepo: storage,
	}
}

func (g *GroupService) CreateTaskGroup(ctx context.Context, name, description string) (dto.TaskGroupDTO, error) {
	taskGroup, err := g.repo.CreateTaskGroup(ctx, name, description)
	if err != nil {
		return dto.TaskGroupDTO{}, err
	}

	dto := g.toDTO(*taskGroup)

	return dto, nil
}

func (g *GroupService) GetTaskGroups(ctx context.Context) ([]dto.TaskGroupDTO, error) {
	taskGroups, err := g.repo.GetTaskGroups(ctx)

	if err != nil {
		return []dto.TaskGroupDTO{}, err
	}

	taskGroupsDTO := make([]dto.TaskGroupDTO, len(taskGroups))
	for i, task := range taskGroups {
		taskGroupsDTO[i] = g.toDTO(task)
	}

	return taskGroupsDTO, nil
}

func (g *GroupService) AssignTaskToGroup(ctx context.Context, taskID, groupID int) error {
	if err := g.repo.AssignTaskToGroup(ctx, taskID, groupID); err != nil {
		return err
	}

	return nil
}

func (g *GroupService) RemoveTaskFromGroup(ctx context.Context, taskID int) error {
	if err := g.repo.RemoveTaskFromGroup(ctx, taskID); err != nil {
		return err
	}

	return nil
}

func (g *GroupService) GetTasksByGroup(ctx context.Context, groupID int, statusFilter *string) ([]dto.TaskDTO, error) {
	tasks, err := g.repo.GetTasksByGroup(ctx, groupID, statusFilter)
	if err != nil {
		return []dto.TaskDTO{}, err
	}

	tasksDTO := make([]dto.TaskDTO, len(tasks))
	for i, task := range tasks {
		tasksDTO[i] = g.taskToDTO(task)
	}

	return tasksDTO, nil
}

func (g *GroupService) toDTO(taskGroup model.TaskGroup) dto.TaskGroupDTO {
	var taskGroupDTO dto.TaskGroupDTO

	taskGroupDTO.ID = taskGroup.ID
	taskGroupDTO.Name = taskGroup.Name
	taskGroupDTO.Description = taskGroup.Description
	taskGroupDTO.CreatedAt = taskGroup.CreatedAt

	return taskGroupDTO
}

func (g *GroupService) taskToDTO(t storage.TaskWithRelations) dto.TaskDTO {
	return dto.TaskDTO{
		ID:              t.Task.ID,
		GroupID:         t.GroupID,
		Username:        t.Username,
		GroupName:       t.GroupName,
		Name:            t.Task.Name,
		Description:     t.Task.Description,
		Author:          t.Task.Author,
		Status:          t.Task.Status,
		SolutionComment: t.Task.SolutionComment,
		CreatedAt:       t.Task.CreatedAt,
		UpdatedAt:       t.Task.UpdatedAt,
		CompletedAt:     t.Task.CompletedAt,
	}
}