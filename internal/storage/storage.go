package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"tasksite/internal/model"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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

func (s *Storage) CreateTask(name string) (*model.Task, error) {
	result, err := s.db.Exec(
		"INSERT INTO tasks (name) VALUES (?)",
		name,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &model.Task{
		ID:        int(id),
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}

func (s *Storage) GetTasks() ([]model.Task, error) {
	rows, err := s.db.Query("SELECT id, name, created_at FROM tasks ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		var created_at string

		if err := rows.Scan(&task.ID, &task.Name, &created_at); err != nil {
			return nil, err
		}

		task.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created_at)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *Storage) DeleteTask(id int) error {
	result, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task with id %d not found", id)
	}

	return nil
}

// Close закрывает подключение к БД
func (s *Storage) Close() error {
	return s.db.Close()
}
