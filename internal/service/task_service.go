package service

import (
	"context"
	"fmt"
	"strings"
	"tasksite/internal/model"
	"tasksite/internal/model/dto"
	"tasksite/internal/storage"
)

type TaskService struct {
	storage *storage.Storage
}

type PaginatedTasksResponse struct {
	Tasks      []dto.TaskDTO        `json:"tasks"`
	Pagination model.PaginationMeta `json:"pagination"`
}

func NewTaskService(storage *storage.Storage) *TaskService {
	return &TaskService{storage: storage}
}

func (s *TaskService) GetTasks(ctx context.Context, statusFilter *string) ([]dto.TaskDTO, error) {
	tasks, err := s.storage.GetTasks(ctx, statusFilter)

	if err != nil {
		return nil, err
	}

	tasksDTO := make([]dto.TaskDTO, len(tasks))
	for i, task := range tasks {
		tasksDTO[i] = s.toDTO(ctx, task)
	}

	return tasksDTO, nil
}

func (s *TaskService) GetTasksPaginated(ctx context.Context, pq model.PaginationQuery, groupID *int) (PaginatedTasksResponse, error) {
    pq.Validate()
    
    total, err := s.storage.CountTasks(ctx, 
        func() *string { if pq.Status != "" { return &pq.Status }; return nil }(), 
        groupID)
    if err != nil {
        return PaginatedTasksResponse{}, fmt.Errorf("failed to count tasks: %w", err)
    }
    
    tasks, err := s.storage.GetTasksPaginated(ctx, pq, groupID)
    if err != nil {
        return PaginatedTasksResponse{}, fmt.Errorf("failed to get tasks: %w", err)
    }
    
    dtos := make([]dto.TaskDTO, len(tasks))
    for i, t := range tasks {
        dtos[i] = s.toDTO(ctx, t)
    }
    
    return PaginatedTasksResponse{
        Tasks:      dtos,
        Pagination: model.NewPaginationMeta(pq.Page, pq.Limit, total),
    }, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int) (dto.TaskDTO, error) {
	task, err := s.storage.GetTaskByID(ctx, id)

	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("failed to get task: %w", err)
	}

	taskDTO := s.toDTO(ctx, *task)

	return taskDTO, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID int, username string) error {
	user, err := s.storage.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("User not found: %w", err)
	}

	if err := s.storage.DeleteTask(ctx, taskID, user.ID); err != nil {
		return fmt.Errorf("Failed to delete task: %w", err)
	}

	return nil
}

func (s *TaskService) ClaimTask(ctx context.Context, taskID int, username string) (dto.TaskDTO, error) {
	user, err := s.storage.GetUserByUsername(ctx, username)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("User not found: %w", err)
	}

	if err := s.storage.ClaimTask(ctx, taskID, user.ID); err != nil {
		if strings.Contains(err.Error(), "already claimed") {
			return dto.TaskDTO{}, fmt.Errorf("Task already claimed: %w", err)
		}
		return dto.TaskDTO{}, fmt.Errorf("Failed to claim task: %w", err)
	}

	task, err := s.storage.GetTaskByID(ctx, taskID)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("Failed to fetch task: %w", err)
	}

	return s.toDTO(ctx, *task), nil
}

func (s *TaskService) CompleteTask(ctx context.Context, taskID int, username string) (dto.TaskDTO, error) {
	user, err := s.storage.GetUserByUsername(ctx, username)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("User not found: %w", err)
	}

	task, err := s.storage.CompleteTask(ctx, taskID, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return dto.TaskDTO{}, fmt.Errorf("Task not found: %w", err)
		}
		if strings.Contains(err.Error(), "forbidden") {
			return dto.TaskDTO{}, fmt.Errorf("forbidden: %w", err)
		}
		return dto.TaskDTO{}, fmt.Errorf("Failed to complete task: %w", err)
	}

	return s.toDTO(ctx, *task), nil
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID int, req model.UpdateTaskRequest, username string) (dto.TaskDTO, error) {
	user, err := s.storage.GetUserByUsername(ctx, username)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("User not found: %w", err)
	}

	task, err := s.storage.UpdateTask(ctx, taskID, req, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") {
			return dto.TaskDTO{}, fmt.Errorf("access denied: %w", err)
		}
		if strings.Contains(err.Error(), "not found") {
			return dto.TaskDTO{}, fmt.Errorf("Task not found: %w", err)
		}
		return dto.TaskDTO{}, fmt.Errorf("Failed to update task: %w", err)
	}

	return s.toDTO(ctx, *task), nil
}

func (s *TaskService) GetUngroupedTasks(ctx context.Context, statusFilter *string) ([]dto.TaskDTO, error) {
	tasks, err := s.storage.GetUngroupedTasks(ctx, statusFilter)
	if err != nil {
		return nil, err
	}
	dtos := make([]dto.TaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = s.toDTO(ctx, t)
	}
	return dtos, nil
}

func (s *TaskService) toDTO(ctx context.Context, task model.Task) dto.TaskDTO {
	var taskDTO dto.TaskDTO
	username := ""
	if task.UserID != nil {
		user, err := s.storage.GetUserById(ctx, *task.UserID)
		if err == nil {
			username = user.Username
		}

	}


	taskDTO.ID = task.ID
	taskDTO.Name = task.Name
	// taskDTO.GroupName = task.GroupID
	taskDTO.Author = task.Author
	taskDTO.CompletedAt = task.CompletedAt
	taskDTO.CreatedAt = task.CreatedAt
	taskDTO.Description = task.Description
	taskDTO.Status = task.Status
	taskDTO.UpdatedAt = task.UpdatedAt
	taskDTO.Username = username

	return taskDTO
}
