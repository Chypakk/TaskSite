package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"tasksite/internal/model"
	"tasksite/internal/storage"
)

type TaskHandler struct {
	storage *storage.Storage
}

func NewTaskHandler(storage *storage.Storage) *TaskHandler {
	return &TaskHandler{storage: storage}
}

// CreateTask godoc
// @Summary      Создать задачу
// @Description  Создаёт новую задачу (требуется авторизация)
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     SessionToken
// @Param        request  body  model.CreateTaskRequest  true  "Название задачи"
// @Success      201  {object}  model.Task
// @Failure      400  {string}  string  "Invalid request body"
// @Failure      401  {string}  string  "Unauthorized"
// @Router       /tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, "Task name is required", http.StatusBadRequest)
		return
	}

	task, err := h.storage.CreateTask(req.Name, req.Description, req.Author)
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetTasks godoc
// @Summary      Получить все задачи
// @Description  Возвращает список задач (требуется авторизация)
// @Tags         tasks
// @Produce      json
// @Security     SessionToken
// @Success      200  {array}  model.Task
// @Failure      401  {string}  string  "Unauthorized"
// @Router       /tasks [get]
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var statusFilter *string

	if status := r.URL.Query().Get("status"); status != "" {
		statusFilter = &status
	}

	tasks, err := h.storage.GetTasks(statusFilter)
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []model.Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	username, ok := r.Context().Value("username").(string)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	user, err := h.storage.GetUserByUsername(username)
    if err != nil {
        http.Error(w, "User not found", http.StatusInternalServerError)
        return
    }

	if err := h.storage.DeleteTask(id, user.ID); err != nil {
		http.Error(w, "Failed to delete task", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) ClaimTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

	username, ok := r.Context().Value("username").(string)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	user, err := h.storage.GetUserByUsername(username)
    if err != nil {
        http.Error(w, "User not found", http.StatusInternalServerError)
        return
    }

	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
    path = strings.TrimSuffix(path, "/claim")
    taskID, err := strconv.Atoi(path)
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }

    if err := h.storage.ClaimTask(taskID, user.ID); err != nil {
        if strings.Contains(err.Error(), "already claimed") {
            http.Error(w, "Task already claimed", http.StatusConflict)
            return
        }
        http.Error(w, "Failed to claim task", http.StatusInternalServerError)
        return
    }

    task, err := h.storage.GetTaskByID(taskID)
    if err != nil {
        http.Error(w, "Failed to fetch task", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}
