package service

import (
	"fmt"
	"strings"
	"tasksite/internal/model"
	"tasksite/internal/model/dto"
	"tasksite/internal/storage"
)

type TaskService struct {
	storage *storage.Storage
}

func NewTaskService(storage *storage.Storage) *TaskService {
	return &TaskService{storage: storage}
}

func (s *TaskService) GetTasks(statusFilter *string) ([]dto.TaskDTO, error) {
	tasks, err := s.storage.GetTasks(statusFilter)

	if err != nil {
		return nil, err
	}

	tasksDTO := make([]dto.TaskDTO, len(tasks))
	for i, task := range tasks {
		tasksDTO[i] = s.toDTO(task)
	}

	return tasksDTO, nil
}

func (s *TaskService) GetTaskByID(id int) (dto.TaskDTO, error) {
	task, err := s.storage.GetTaskByID(id)

	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("failed to get task: %w", err)
	}

	taskDTO := s.toDTO(*task)

	return taskDTO, nil
}

func (s *TaskService) DeleteTask(taskID int, username string) error {
	user, err := s.storage.GetUserByUsername(username)
	if err != nil {
		return fmt.Errorf("User not found: %w", err)
	}

	if err := s.storage.DeleteTask(taskID, user.ID); err != nil {
		return fmt.Errorf("Failed to delete task: %w", err)
	}

	return nil
}

func (s *TaskService) ClaimTask(taskID int, username string) (dto.TaskDTO, error) {
	user, err := s.storage.GetUserByUsername(username)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("User not found: %w", err)
	}

	if err := s.storage.ClaimTask(taskID, user.ID); err != nil {
		if strings.Contains(err.Error(), "already claimed") {
			return dto.TaskDTO{}, fmt.Errorf("Task already claimed: %w", err)
		}
		return dto.TaskDTO{}, fmt.Errorf("Failed to claim task: %w", err)
	}

	task, err := s.storage.GetTaskByID(taskID)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("Failed to fetch task: %w", err)
	}

	return s.toDTO(*task), nil
}

func (s *TaskService) CompleteTask(taskID int, username string) (dto.TaskDTO, error) {
	user, err := s.storage.GetUserByUsername(username)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("User not found: %w", err)
	}

	task, err := s.storage.CompleteTask(taskID, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return dto.TaskDTO{}, fmt.Errorf("Task not found: %w", err)
		}
		if strings.Contains(err.Error(), "forbidden") {
			return dto.TaskDTO{}, fmt.Errorf("forbidden: %w", err)
		}
		return dto.TaskDTO{}, fmt.Errorf("Failed to complete task: %w", err)
	}

	return s.toDTO(*task), nil
}

func (s *TaskService) UpdateTask(taskID int, req model.UpdateTaskRequest, username string) (dto.TaskDTO, error) {
	user, err := s.storage.GetUserByUsername(username)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("User not found: %w", err)
	}

	task, err := s.storage.UpdateTask(taskID, req, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") {
			return dto.TaskDTO{}, fmt.Errorf("access denied: %w", err)
		}
		if strings.Contains(err.Error(), "not found") {
			return dto.TaskDTO{}, fmt.Errorf("Task not found: %w", err)
		}
		return dto.TaskDTO{}, fmt.Errorf("Failed to update task: %w", err)
	}

	return s.toDTO(*task), nil
}

func (s *TaskService) toDTO(task model.Task) dto.TaskDTO {
	var taskDTO dto.TaskDTO
	username := ""
	if task.UserID != nil {
		user, err := s.storage.GetUserById(*task.UserID)
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
