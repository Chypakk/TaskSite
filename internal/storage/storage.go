package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strings"
	"tasksite/internal/logger"
	"tasksite/internal/model"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Storage struct {
	db *sql.DB
}

func ConnectDB(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &Storage{db: db}

	if err := runMigrations(storage.db); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return storage, nil
}

func (s *Storage) Ping() error {
	return s.db.Ping()
}

func runMigrations(db *sql.DB) error {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create sqlite driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"sqlite",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}

// --- РАБОТА С ТАСКАМИ ---

func (s *Storage) CreateTask(ctx context.Context, name, description, author string) (*model.Task, error) {
	start := time.Now()

	result, err := s.db.Exec(
		"INSERT INTO tasks (name, description, author, status) VALUES (?, ?, ?, 'open')",
		name,
		description,
		author,
	)

	duration := time.Since(start)
	s.logDBOp(ctx, "create_task", duration, err, "name", name, "author", author)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &model.Task{
		ID:          int(id),
		Name:        name,
		Author:      author,
		Status:      "open",
		CreatedAt:   time.Now(),
		Description: description,
	}, nil
}

func (s *Storage) GetTasks(ctx context.Context, statusFilter *string) ([]model.Task, error) {
	start := time.Now()

	query := "SELECT id, user_id, name, description, author, status, group_id, solution_comment, created_at, updated_at, completed_at FROM tasks WHERE 1=1"
	var args []any

	if statusFilter != nil {
		query += " AND status = ?"
		args = append(args, *statusFilter)
	}

	query += " ORDER BY id DESC"
	rows, err := s.db.Query(query, args...)

	duration := time.Since(start)
	s.logDBOp(ctx, "get_tasks", duration, err, "status_filter", statusFilter)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
//
	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		var createdAtStr, completedAtStr, updatedAtStr sql.NullString

		if err := rows.Scan(&task.ID, &task.UserID, &task.Name, &task.Description, &task.Author, &task.Status, &task.GroupID, &task.SolutionComment, &createdAtStr, &updatedAtStr, &completedAtStr); err != nil {
			return nil, err
		}

		if createdAtStr.Valid {
			task.CreatedAt, err = parseTime(createdAtStr)
			if err != nil {
				return nil, fmt.Errorf("parse created_at: %w", err)
			}
		}

		if completedAtStr.Valid {
			task.CompletedAt, err = parseTime(completedAtStr)
			if err != nil {
				return nil, fmt.Errorf("parse completed_at: %w", err)
			}
		}

		if updatedAtStr.Valid {
			task.UpdatedAt, err = parseTime(updatedAtStr)
			if err != nil {
				return nil, fmt.Errorf("parse completed_at: %w", err)
			}
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *Storage) GetTaskByID(ctx context.Context, taskID int) (*model.Task, error) {
	start := time.Now()

	row := s.db.QueryRow(
		"SELECT id, user_id, name, description, author, status, created_at, updated_at, completed_at FROM tasks WHERE id = ?",
		taskID,
	)

	var task model.Task
	var createdAtStr, completedAtStr, updatedAtStr sql.NullString

	err := row.Scan(&task.ID, &task.UserID, &task.Name, &task.Description, &task.Author, &task.Status, &createdAtStr, &updatedAtStr, &completedAtStr)

	duration := time.Since(start)
	s.logDBOp(ctx, "get_task_by_id", duration, err, "task_id", taskID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}

	if createdAtStr.Valid {
		task.CreatedAt, err = parseTime(createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
	}

	if completedAtStr.Valid {
		task.CompletedAt, err = parseTime(completedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse completed_at: %w", err)
		}
	}
	if updatedAtStr.Valid {
		task.UpdatedAt, err = parseTime(updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse completed_at: %w", err)
		}
	}

	return &task, nil
}

func (s *Storage) ClaimTask(ctx context.Context, taskId, userId int) error {
	start := time.Now()

	result, err := s.db.Exec(
		"UPDATE tasks SET user_id = ?, status = 'in_progress' WHERE id = ? AND status != 'completed'",
		userId, taskId,
	)

	duration := time.Since(start)
	s.logDBOp(ctx, "claim_task", duration, err, "task_id", taskId, "user_id", userId)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task not found or already claimed")
	}

	return nil
}

func (s *Storage) CompleteTask(ctx context.Context, taskID, userID int) (*model.Task, error) {
	start := time.Now()
	row := s.db.QueryRow(
		`SELECT id, user_id, name, description, author, status, created_at, updated_at, completed_at 
         FROM tasks WHERE id = ? AND user_id = ? AND status = 'in_progress'`,
		taskID, userID,
	)

	var task model.Task
	var createdAtStr, completedAtStr, updatedAtStr sql.NullString
	err := row.Scan(&task.ID, &task.UserID, &task.Name, &task.Description,
		&task.Author, &task.Status, &createdAtStr, &updatedAtStr, &completedAtStr)

	duration := time.Since(start)
	s.logDBOp(ctx, "complete_task", duration, err, "task_id", taskID, "user_id", userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found or forbidden")
		}
		return nil, err
	}

	now := time.Now()
	_, err = s.db.Exec(
		"UPDATE tasks SET status = 'completed', completed_at = ? WHERE id = ?",
		now, taskID,
	)
	if err != nil {
		return nil, err
	}

	task.Status = "completed"
	task.CompletedAt = now
	return &task, nil
}

func (s *Storage) DeleteTask(ctx context.Context, id, userId int) error {
	start := time.Now()

	result, err := s.db.Exec("DELETE FROM tasks WHERE id = ? AND (user_id = ? OR user_id IS NULL)", id, userId)

	duration := time.Since(start)
	s.logDBOp(ctx, "delete_task", duration, err, "task_id", id, "user_id", userId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found or access denied")
	}

	return nil
}

func (s *Storage) UpdateTask(ctx context.Context, taskID int, req model.UpdateTaskRequest, editorID int) (*model.Task, error) {
	start := time.Now()

	var task model.Task
	var createdAtStr, completedAtStr sql.NullString
	err := s.db.QueryRow(
		"SELECT id, user_id, name, description, author, status, created_at, completed_at FROM tasks WHERE id = ?",
		taskID,
	).Scan(&task.ID, &task.UserID, &task.Name, &task.Description,
		&task.Author, &task.Status, &createdAtStr, &completedAtStr)

	duration := time.Since(start)
	s.logDBOp(ctx, "update_task", duration, err, "task_id", taskID, "user_id", editorID)

	if err != nil {
		return nil, fmt.Errorf("task not found")
	}

	if task.UserID != nil && *task.UserID != editorID {
		return nil, fmt.Errorf("access denied")
	}

	updates := []string{}
	args := []any{}

	if req.Name != "" {
		updates = append(updates, "name = ?")
		args = append(args, strings.TrimSpace(req.Name))
	}
	if req.Description != "" {
		updates = append(updates, "description = ?")
		args = append(args, req.Description)
	}
	if req.Author != "" {
		updates = append(updates, "author = ?")
		args = append(args, req.Author)
	}
	if req.SolutionComment != "" {
		updates = append(updates, "solution_comment = ?")
		args = append(args, req.SolutionComment)
	}
	if req.Status != "" {
		valid := map[string]bool{
			"open": true, "in_progress": true, "completed": true, "closed": true,
			"has_solution": true,
	}
		if !valid[req.Status] {
			return nil, fmt.Errorf("invalid status value")
		}
		updates = append(updates, "status = ?")
		args = append(args, req.Status)

		// Авто-заполнение completed_at при смене статуса
		if req.Status == "completed" {
			updates = append(updates, "completed_at = COALESCE(completed_at, ?)")
			args = append(args, time.Now())
		} else {
			updates = append(updates, "completed_at = NULL")
		}

		if req.Status == "open" {
			updates = append(updates, "user_id = null")
		}
	}

	if len(updates) == 0 {
		return &task, nil // Нечего обновлять
	}

	args = append(args, taskID)

	query := fmt.Sprintf("UPDATE tasks SET %s, updated_at = CURRENT_TIMESTAMP WHERE id = ?", strings.Join(updates, ", "))

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return s.GetTaskByID(ctx, taskID)
}

func (s *Storage) GetUngroupedTasks(ctx context.Context, statusFilter *string) ([]model.Task, error) {
    start := time.Now()
    query := `SELECT id, user_id, name, description, author, status, group_id, 
                     solution_comment, created_at, updated_at, completed_at 
              FROM tasks WHERE group_id IS NULL`
    var args []any
    
    if statusFilter != nil {
        query += " AND status = ?"
        args = append(args, *statusFilter)
    }
    query += " ORDER BY created_at DESC"
    
    rows, err := s.db.Query(query, args...)
    duration := time.Since(start)
    s.logDBOp(ctx, "get_ungrouped_tasks", duration, err, "status_filter", statusFilter)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		var createdAtStr, completedAtStr, updatedAtStr sql.NullString

		if err := rows.Scan(&task.ID, &task.UserID, &task.Name, &task.Description, &task.Author, &task.Status, &createdAtStr, &updatedAtStr, &completedAtStr); err != nil {
			return nil, err
		}

		if createdAtStr.Valid {
			task.CreatedAt, err = parseTime(createdAtStr)
			if err != nil {
				return nil, fmt.Errorf("parse created_at: %w", err)
			}
		}

		if completedAtStr.Valid {
			task.CompletedAt, err = parseTime(completedAtStr)
			if err != nil {
				return nil, fmt.Errorf("parse completed_at: %w", err)
			}
		}

		if updatedAtStr.Valid {
			task.UpdatedAt, err = parseTime(updatedAtStr)
			if err != nil {
				return nil, fmt.Errorf("parse completed_at: %w", err)
			}
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Close закрывает подключение к БД
func (s *Storage) Close() error {
	return s.db.Close()
}

// --- РАБОТА С ПОЛЬЗОВАТЕЛЯМИ ---

func (s *Storage) CreateUser(ctx context.Context, username, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	res, err := s.db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, hash)

	duration := time.Since(start)
    s.logDBOp(ctx, "create_user", duration, err, "username", username)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, fmt.Errorf("username already exists")
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:           int(id),
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}, nil
}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	start := time.Now()
	
	row := s.db.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = ?",
		username,
	)

	var user model.User
	var createdAtStr sql.NullString

	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &createdAtStr)

	duration := time.Since(start)
    s.logDBOp(ctx, "get_user_by_username", duration, err, "username", username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	user.CreatedAt, _ = parseTime(createdAtStr)

	return &user, nil
}

func (s *Storage) GetUserById(ctx context.Context, id int) (*model.User, error) {
	row := s.db.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE id = ?",
		id,
	)

	var user model.User
	var createdAtStr sql.NullString

	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	user.CreatedAt, _ = parseTime(createdAtStr)

	return &user, nil
}

// --- РАБОТА С ГРУППАМИ ЗАДАЧ ---

func (s *Storage) CreateTaskGroup(ctx context.Context, name, description string) (*model.TaskGroup, error) {
	start := time.Now()
    res, err := s.db.Exec(
        "INSERT INTO task_groups (name, description) VALUES (?, ?)",
        name, description,
    )
    duration := time.Since(start)
    s.logDBOp(ctx, "create_task_group", duration, err, "name", name)
	if err != nil {
        if strings.Contains(err.Error(), "UNIQUE constraint failed") {
            return nil, fmt.Errorf("group with this name already exists")
        }
        return nil, err
    }
    id, _ := res.LastInsertId()
    return &model.TaskGroup{
        ID:          int(id),
        Name:        name,
        Description: description,
        CreatedAt:   time.Now(),
    }, nil
}

func (s *Storage) GetTaskGroups(ctx context.Context) ([]model.TaskGroup, error) {
	start := time.Now()
    rows, err := s.db.Query("SELECT id, name, description, created_at FROM task_groups ORDER BY name")
    duration := time.Since(start)
    s.logDBOp(ctx, "get_task_groups", duration, err)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var groups []model.TaskGroup
    for rows.Next() {
        var g model.TaskGroup
        var createdAtStr sql.NullString
        if err := rows.Scan(&g.ID, &g.Name, &g.Description, &createdAtStr); err != nil {
            return nil, err
        }
        if createdAtStr.Valid {
            g.CreatedAt, _ = parseTime(createdAtStr)
        }
        groups = append(groups, g)
    }
    return groups, nil
}

func (s *Storage) AssignTaskToGroup(ctx context.Context, taskID, groupID int) error {
	start := time.Now()
    var exists int
    err := s.db.QueryRow("SELECT COUNT(*) FROM task_groups WHERE id = ?", groupID).Scan(&exists)
    if err != nil || exists == 0 {
        return fmt.Errorf("group not found")
    }
    
    _, err = s.db.Exec("UPDATE tasks SET group_id = ? WHERE id = ?", groupID, taskID)
    duration := time.Since(start)
    s.logDBOp(ctx, "assign_task_to_group", duration, err, "task_id", taskID, "group_id", groupID)
    return err
}

func (s *Storage) RemoveTaskFromGroup(ctx context.Context, taskID int) error {
	start := time.Now()
    _, err := s.db.Exec("UPDATE tasks SET group_id = NULL WHERE id = ?", taskID)
    duration := time.Since(start)
    s.logDBOp(ctx, "remove_task_from_group", duration, err, "task_id", taskID)
    return err
}

func (s *Storage) GetTasksByGroup(ctx context.Context, groupID int, statusFilter *string) ([]model.Task, error) {
	start := time.Now()
    query := `SELECT id, user_id, name, description, author, status, group_id, created_at, updated_at, completed_at 
              FROM tasks WHERE group_id = ?`
    var args []any = []any{groupID}
    
    if statusFilter != nil {
        query += " AND status = ?"
        args = append(args, *statusFilter)
    }
    query += " ORDER BY created_at DESC"
    
    rows, err := s.db.Query(query, args...)
    duration := time.Since(start)
    s.logDBOp(ctx, "get_tasks_by_group", duration, err, "group_id", groupID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		var createdAtStr, completedAtStr, updatedAtStr sql.NullString

		if err := rows.Scan(&task.ID, &task.UserID, &task.Name, &task.Description, &task.Author, &task.Status, &createdAtStr, &updatedAtStr, &completedAtStr); err != nil {
			return nil, err
		}

		if createdAtStr.Valid {
			task.CreatedAt, err = parseTime(createdAtStr)
			if err != nil {
				return nil, fmt.Errorf("parse created_at: %w", err)
			}
		}

		if completedAtStr.Valid {
			task.CompletedAt, err = parseTime(completedAtStr)
			if err != nil {
				return nil, fmt.Errorf("parse completed_at: %w", err)
			}
		}

		if updatedAtStr.Valid {
			task.UpdatedAt, err = parseTime(updatedAtStr)
			if err != nil {
				return nil, fmt.Errorf("parse completed_at: %w", err)
			}
		}

		tasks = append(tasks, task)
	}
    return tasks, nil
}

func parseTime(nullStr sql.NullString) (time.Time, error) {
	if !nullStr.Valid {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, nullStr.String)
}

func (s *Storage) logDBOp(ctx context.Context, opName string, duration time.Duration, err error, args ...any) {
	log := logger.FromContext(ctx)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("DB: %s failed", opName), err, args...)
	} else {
		log.Info(ctx, fmt.Sprintf("DB: %s", opName), append([]any{"duration", duration}, args...)...)
	}
}
