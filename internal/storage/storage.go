package storage

import (
	"context"
	"database/sql"
	"fmt"
	"tasksite/internal/config"
	"tasksite/internal/logger"
	"tasksite/internal/model"
	"time"
)

type Storage interface{
	// --- РАБОТА С ТАСКАМИ ---

	GetTasks(ctx context.Context, taskID, groupID *int, limit, offset int, statusFilter *string, sortBy string) ([]TaskWithRelations, error)
	CreateTask(ctx context.Context, name, description, author string) (*model.Task, error)
	ClaimTask(ctx context.Context, taskId, userId int) error
	CompleteTaskOld(ctx context.Context, taskID, userID int) (*model.Task, error)
	CompleteTask(ctx context.Context, taskID, userID int) error
	DeleteTask(ctx context.Context, id, userId int) error
	UpdateTask(ctx context.Context, taskID int, req model.UpdateTaskRequest, editorID int) (*TaskWithRelations, error)
	CountTasks(ctx context.Context, statusFilter *string, groupID *int) (int, error)
	
	// --- РАБОТА С ПОЛЬЗОВАТЕЛЯМИ ---

	CreateUser(ctx context.Context, username, password string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserById(ctx context.Context, id int) (*model.User, error)
	
	// --- РАБОТА С ГРУППАМИ ЗАДАЧ ---

	CreateTaskGroup(ctx context.Context, name, description string) (*model.TaskGroup, error)
	GetTaskGroups(ctx context.Context) ([]model.TaskGroup, error)
	GetTaskGroupById(ctx context.Context, id int) (*model.TaskGroup, error)
	AssignTaskToGroup(ctx context.Context, taskID, groupID int) error
	RemoveTaskFromGroup(ctx context.Context, taskID int) error
	GetTasksByGroup(ctx context.Context, groupID int, statusFilter *string) ([]TaskWithRelations, error)
	EditGroup(ctx context.Context, groupID int, name, description string) error
	
	// --- РАБОТА С СЕССИЯМИ ---
	
	CreateSession(ctx context.Context, token, username string, expiresAt time.Time) error
	GetSessionByToken(ctx context.Context, token string) (*model.Session, error)
	UpdateSessionExpires(ctx context.Context, token string, newExpiredAt time.Time) error
	DeleteSession(ctx context.Context, token string) error
	
	Close() error
	Ping() error
}

func ConnectDB(cfg config.Config) (Storage, error) {
	switch cfg.DBDriver {
	case "postgres":
		return newPostgresStorage(cfg) // из storage_postgres.go
	case "sqlite":
		return newSqliteStorage(cfg) // из storage_sqlite.go
	default:
		return nil, fmt.Errorf("unknown driver: %s", cfg.DBDriver)
	}
}

func logDBOp(ctx context.Context, opName string, duration time.Duration, err error, args ...any) {
	log := logger.FromContext(ctx)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("DB: %s failed", opName), err, args...)
	} else {
		log.Info(ctx, fmt.Sprintf("DB: %s", opName), append([]any{"duration", duration}, args...)...)
	}
}

func scanTasks(rows *sql.Rows) ([]model.Task, error) {
	tasks := make([]model.Task, 0)
	for rows.Next() {
		var t model.Task
		var createdAtStr, updatedAtStr, completedAtStr sql.NullString

		err := rows.Scan(
			&t.ID, &t.UserID, &t.Name, &t.Description, &t.Author,
			&t.Status, &t.GroupID, &t.SolutionComment,
			&createdAtStr, &updatedAtStr, &completedAtStr,
		)
		if err != nil {
			return nil, err
		}

		if createdAtStr.Valid {
			t.CreatedAt, _ = parseTime(createdAtStr)
		}
		if updatedAtStr.Valid {
			t.UpdatedAt, _ = parseTime(updatedAtStr)
		}
		if completedAtStr.Valid {
			t.CompletedAt, _ = parseTime(completedAtStr)
		}

		tasks = append(tasks, t)
	}
	return tasks, nil
}

func scanTasksWithRelations(rows *sql.Rows) ([]TaskWithRelations, error) {
	result := make([]TaskWithRelations, 0)
	for rows.Next() {
		var t TaskWithRelations
		var createdAtStr, updatedAtStr, completedAtStr, username, groupName, groupDesc sql.NullString

		err := rows.Scan(
			&t.Task.ID, &t.Task.UserID, &t.Task.Name, &t.Task.Description, &t.Task.Author,
			&t.Task.Status, &t.Task.GroupID, &t.Task.SolutionComment,
			&createdAtStr, &updatedAtStr, &completedAtStr,
			&username, &groupName, &groupDesc,
		)
		if err != nil {
			return nil, err
		}

		if createdAtStr.Valid {
			t.Task.CreatedAt, _ = parseTime(createdAtStr)
		}
		if updatedAtStr.Valid {
			t.Task.UpdatedAt, _ = parseTime(updatedAtStr)
		}
		if completedAtStr.Valid {
			t.Task.CompletedAt, _ = parseTime(completedAtStr)
		}
		if username.Valid {
			t.Username = username.String
		}
		if groupName.Valid {
			t.GroupName = groupName.String
		}
		if groupDesc.Valid {
			t.GroupDesc = groupDesc.String
		}

		result = append(result, t)
	}
	return result, nil
}