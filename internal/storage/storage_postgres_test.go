package storage

import (
	"context"
	"os"
	"path/filepath"
	"tasksite/internal/config"
	"testing"

	"github.com/joho/godotenv"
)

func getTestConfig() config.Config {
	envPath := findEnvFile()

	if err := godotenv.Load(envPath); err != nil {
		return config.Config{}
	}

	return config.Config{
		DBDriver: "postgres",
		DBHost:   getEnv("DB_HOST"),
		DBPort:   getEnv("DB_PORT"),
		DBUser:   getEnv("DB_USER"),
		DBPass:   getEnv("DB_PASS"),
		DBName:   getEnv("DB_NAME"),
	}
}

func getEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return ""
}

func cleanDB(t *testing.T, s *postgresStorage) {
	t.Helper()
	// TRUNCATE быстрее DELETE и сбрасывает автоинкремент
	_, err := s.db.Exec(`
		TRUNCATE TABLE sessions, task_groups, tasks, users RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Logf("Warning: cleanup failed: %v", err)
	}
}

func findEnvFile() string {
	// Рядом с exe
	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Dir(exePath)
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	// В текущей рабочей директории
	if cwd, err := os.Getwd(); err == nil {
		envPath := filepath.Join(cwd, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	return ""
}

// Тест 1: Регистрация и вход
func TestPostgres_UserRegisterLogin(t *testing.T) {
	cfg := getTestConfig()
	s, err := newPostgresStorage(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}
	defer s.Close()
	defer cleanDB(t, s.(*postgresStorage)) // Чистим базу после теста

	ctx := context.Background()

	// 1. Создаём пользователя
	created, err := s.CreateUser(ctx, "testuser", "password123")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if created.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", created.Username)
	}
	if created.ID <= 0 {
		t.Error("Expected positive ID after insert")
	}

	// 2. Получаем его по имени (имитация логина)
	fetched, err := s.GetUserByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}

	// 3. Проверяем, что пароль хеширован (bcrypt всегда разный, сравнивать нельзя)
	// Просто убеждаемся, что хеш есть и он не пустой
	if fetched.PasswordHash == "" {
		t.Error("Password hash is empty")
	}
	if fetched.ID != created.ID {
		t.Errorf("ID mismatch: expected %d, got %d", created.ID, fetched.ID)
	}

	// 4. Проверяем дубликат
	_, err = s.CreateUser(ctx, "testuser", "another_pass")
	if err == nil {
		t.Error("Expected error on duplicate username, got nil")
	}
}

// Тест 2: Жизненный цикл задачи
func TestPostgres_TaskLifecycle(t *testing.T) {
	cfg := getTestConfig()
	s, err := newPostgresStorage(cfg)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer s.Close()
	defer cleanDB(t, s.(*postgresStorage))

	ctx := context.Background()

	user, err := s.CreateUser(ctx, "test_dev", "password123")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// 1. Создаём задачу
	task, err := s.CreateTask(ctx, "Fix login bug", "Users can't login", "admin")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.Status != "open" {
		t.Errorf("Expected status 'open', got '%s'", task.Status)
	}

	// 2. Получаем список задач
	tasks, err := s.GetTasks(ctx, nil, nil, 0, 0, nil, "")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// 3. Берём в работу
	err = s.ClaimTask(ctx, task.ID, user.ID)
	if err != nil {
		t.Fatalf("ClaimTask failed: %v", err)
	}

	// 4. Завершаем задачу
	err = s.CompleteTask(ctx, task.ID, user.ID)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// 5. Проверяем, что статус изменился
	updated, err := s.GetTasks(ctx, &task.ID, nil, 0, 0, nil, "")
	if err != nil {
		t.Fatalf("GetTasks by ID failed: %v", err)
	}
	if len(updated) == 0 || updated[0].Status != "completed" {
		t.Errorf("Expected completed task, got %+v", updated)
	}
}

// Тест 3: Группы задач
func TestPostgres_TaskGroups(t *testing.T) {
	cfg := getTestConfig()
	s, err := newPostgresStorage(cfg)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer s.Close()
	defer cleanDB(t, s.(*postgresStorage))

	ctx := context.Background()

	// 1. Создаём группу
	group, err := s.CreateTaskGroup(ctx, "Backend Team", "API and core")
	if err != nil {
		t.Fatalf("CreateTaskGroup failed: %v", err)
	}

	// 2. Создаём задачу и привязываем к группе
	task, _ := s.CreateTask(ctx, "Refactor auth", "", "dev")
	err = s.AssignTaskToGroup(ctx, task.ID, group.ID)
	if err != nil {
		t.Fatalf("AssignTaskToGroup failed: %v", err)
	}

	// 3. Получаем задачи группы
	grpTasks, err := s.GetTasksByGroup(ctx, group.ID, nil)
	if err != nil {
		t.Fatalf("GetTasksByGroup failed: %v", err)
	}
	if len(grpTasks) != 1 || grpTasks[0].ID != task.ID {
		t.Errorf("Expected task %d in group, got %+v", task.ID, grpTasks)
	}

	// 4. Проверяем, что в группе есть имя (проверка джойна)
	if grpTasks[0].GroupName != group.Name {
		t.Errorf("Expected group name '%s', got '%s'", group.Name, grpTasks[0].GroupName)
	}
}
