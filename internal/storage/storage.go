package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strings"
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

func (s *Storage) CreateTask(name, description, author string) (*model.Task, error) {
	result, err := s.db.Exec(
		"INSERT INTO tasks (name, description, author, status) VALUES (?, ?, ?, 'pool')",
		name,
		description,
		author,
	)
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
		Status:      "pool",
		CreatedAt:   time.Now(),
		Description: description,
	}, nil
}

func (s *Storage) GetTasks(statusFilter *string) ([]model.Task, error) {
	query := "SELECT id, user_id, name, description, author, status, created_at, completed_at FROM tasks WHERE 1=1"
	var args []any

	if statusFilter != nil {
		query += " AND status = ?"
		args = append(args, *statusFilter)
	}

	query += " ORDER BY id DESC"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		var createdAtStr, completedAtStr sql.NullString

		if err := rows.Scan(&task.ID, &task.UserID, &task.Name, &task.Description, &task.Author, &createdAtStr, &completedAtStr); err != nil {
			return nil, err
		}

		if createdAtStr.Valid {
			task.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr.String)
		}
		if completedAtStr.Valid {
			task.CompletedAt, _ = time.Parse("2006-01-02 15:04:05", completedAtStr.String)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *Storage) GetTaskByID(taskID int) (*model.Task, error) {
	row := s.db.QueryRow(
		"SELECT id, user_id, name, description, author, status, created_at, completed_at FROM tasks WHERE id = ?",
		taskID,
	)

	var task model.Task
	var createdAtStr, completedAtStr sql.NullString

	err := row.Scan(&task.ID, &task.UserID, &task.Name, &task.Description, &task.Author, &createdAtStr, &completedAtStr)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}

	if createdAtStr.Valid {
		task.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr.String)
	}
	if completedAtStr.Valid {
		task.CompletedAt, _ = time.Parse("2006-01-02 15:04:05", completedAtStr.String)
	}

	return &task, nil
}

func (s *Storage) ClaimTask(taskId, userId int) error {
	result, err := s.db.Exec(
		"UPDATE tasks SET user_id = ?, status = 'in_progress' WHERE id = ? AND status = 'pool'",
		userId, taskId,
	)

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

func (s *Storage) CompleteTask(taskID, userID int) (any, error) {
	row := s.db.QueryRow(
        `SELECT id, user_id, name, description, author, status, created_at, completed_at 
         FROM tasks WHERE id = ? AND user_id = ? AND status = 'in_progress'`,
        taskID, userID,
    )

	var task model.Task
    var createdAtStr, completedAtStr sql.NullString
    err := row.Scan(&task.ID, &task.UserID, &task.Name, &task.Description, 
                    &task.Author, &task.Status, &createdAtStr, &completedAtStr)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("task not found or forbidden")
        }
        return nil, err
    }

    // Апдейтим статус и время
    now := time.Now()
    _, err = s.db.Exec(
        "UPDATE tasks SET status = 'completed', completed_at = ? WHERE id = ?",
        now.Format("2006-01-02 15:04:05"), taskID,
    )
    if err != nil {
        return nil, err
    }

    task.Status = "completed"
    task.CompletedAt = now
    return &task, nil
}

func (s *Storage) DeleteTask(id, userId int) error {
	result, err := s.db.Exec("DELETE FROM tasks WHERE id = ? AND (user_id = ? OR user_id IS NULL)", id, userId)
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

// Close закрывает подключение к БД
func (s *Storage) Close() error {
	return s.db.Close()
}

// --- РАБОТА С ПОЛЬЗОВАТЕЛЯМИ ---

func (s *Storage) CreateUser(username, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	res, err := s.db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, hash)
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

func (s *Storage) GetUserByUsername(username string) (*model.User, error) {
	row := s.db.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = ?",
		username,
	)

	var user model.User
	var createdAtStr string

	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)

	return &user, nil
}
