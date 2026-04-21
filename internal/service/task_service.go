package service

import (
	"context"
	"fmt"
	"strings"
	"tasksite/internal/model"
	"tasksite/internal/model/dto"
	"tasksite/internal/repository"
	"tasksite/internal/storage"
)

type TaskService struct {
	taskRepo      repository.TaskRepository
	userRepo      repository.UserRepository
	taskGroupRepo repository.TaskGroupRepository
}

type PaginatedTasksResponse struct {
	Tasks      []dto.TaskDTO        `json:"tasks"`
	Pagination model.PaginationMeta `json:"pagination"`
}

func NewTaskService(repo repository.TaskRepository, userRepo repository.UserRepository, taskGroupRepo repository.TaskGroupRepository) *TaskService {
	return &TaskService{
		taskRepo:      repo,
		userRepo:      userRepo,
		taskGroupRepo: taskGroupRepo,
	}
}

func (s *TaskService) GetTasks(ctx context.Context, statusFilter *string) ([]dto.TaskDTO, error) {
	tasks, err := s.taskRepo.GetTasks(ctx, nil, nil, 0, 0, statusFilter, "")

	if err != nil {
		return nil, err
	}

	tasksDTO := make([]dto.TaskDTO, len(tasks))
	for i, task := range tasks {
		tasksDTO[i] = s.taskWithRelationsToDTO(task)
	}

	return tasksDTO, nil
}

func (s *TaskService) GetTasksPaginated(ctx context.Context, pq model.PaginationQuery, groupID *int) (PaginatedTasksResponse, error) {
	pq.Validate()

	total, err := s.taskRepo.CountTasks(ctx,
		func() *string {
			if pq.Status != "" {
				return &pq.Status
			}
			return nil
		}(),
		groupID)
	if err != nil {
		return PaginatedTasksResponse{}, fmt.Errorf("failed to count tasks: %w", err)
	}

	var gid int
	if groupID != nil {
		gid = *groupID
	}

	tasks, err := s.taskRepo.GetTasks(ctx, nil, &gid, pq.Limit, pq.Offset(), nil, pq.Sort)
	if err != nil {
		return PaginatedTasksResponse{}, fmt.Errorf("failed to get tasks: %w", err)
	}

	dtos := make([]dto.TaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = s.taskWithRelationsToDTO(t)
	}

	return PaginatedTasksResponse{
		Tasks:      dtos,
		Pagination: model.NewPaginationMeta(pq.Page, pq.Limit, total),
	}, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int) (dto.TaskDTO, error) {
	task, err := s.taskRepo.GetTasks(ctx, &id, nil, 0, 0, nil, "")

	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("failed to get task: %w", err)
	}

	taskDTO := s.taskWithRelationsToDTO(task[0])

	return taskDTO, nil
}

func (s *TaskService) CreateTask(ctx context.Context, name, description, author string) (dto.TaskDTO, error) {
	task, err := s.taskRepo.CreateTask(ctx, name, description, author)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("error create task: %w", err)
	}

	return s.taskWithRelationsToDTO(storage.TaskWithRelations{
		Task: *task,
	}), nil
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID int, username string) error {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("User not found: %w", err)
	}

	if err := s.taskRepo.DeleteTask(ctx, taskID, user.ID); err != nil {
		return fmt.Errorf("Failed to delete task: %w", err)
	}

	return nil
}

func (s *TaskService) ClaimTask(ctx context.Context, taskID int, username string) (dto.TaskDTO, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("User not found: %w", err)
	}

	if err := s.taskRepo.ClaimTask(ctx, taskID, user.ID); err != nil {
		if strings.Contains(err.Error(), "already claimed") {
			return dto.TaskDTO{}, fmt.Errorf("Task already claimed: %w", err)
		}
		return dto.TaskDTO{}, fmt.Errorf("Failed to claim task: %w", err)
	}

	task, err := s.taskRepo.GetTasks(ctx, &taskID, nil, 0, 0, nil, "")
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("Failed to fetch task: %w", err)
	}

	return s.taskWithRelationsToDTO(task[0]), nil
}

func (s *TaskService) CompleteTask(ctx context.Context, taskID int, username string) error {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("User not found: %w", err)
	}

	//task, err := s.taskRepo.CompleteTask(ctx, taskID, user.ID)
	err = s.taskRepo.CompleteTask(ctx, taskID, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("Task not found: %w", err)
		}
		if strings.Contains(err.Error(), "forbidden") {
			return fmt.Errorf("forbidden: %w", err)
		}
		return fmt.Errorf("Failed to complete task: %w", err)
	}

	//return s.toDTO(ctx, *task), nil
	return nil
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID int, req model.UpdateTaskRequest, username string) (dto.TaskDTO, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("User not found: %w", err)
	}

	task, err := s.taskRepo.UpdateTask(ctx, taskID, req, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") {
			return dto.TaskDTO{}, fmt.Errorf("access denied: %w", err)
		}
		if strings.Contains(err.Error(), "not found") {
			return dto.TaskDTO{}, fmt.Errorf("Task not found: %w", err)
		}
		return dto.TaskDTO{}, fmt.Errorf("Failed to update task: %w", err)
	}

	return s.taskWithRelationsToDTO(*task), nil
}

func (s *TaskService) GetUngroupedTasks(ctx context.Context, statusFilter *string) ([]dto.TaskDTO, error) {
	nonGroup := 0
	tasks, err := s.taskRepo.GetTasks(ctx, nil, &nonGroup, 0, 0, statusFilter, "")
	if err != nil {
		return nil, err
	}
	dtos := make([]dto.TaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = s.taskWithRelationsToDTO(t)
	}
	return dtos, nil
}

func (s *TaskService) taskWithRelationsToDTO(t storage.TaskWithRelations) dto.TaskDTO {
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
