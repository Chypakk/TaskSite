package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"tasksite/internal/model"
	"tasksite/internal/service"
	"tasksite/internal/storage"
)

type TaskHandler struct {
	taskService *service.TaskService
	storage *storage.Storage
}

func NewTaskHandler(storage *storage.Storage) *TaskHandler {
	return &TaskHandler{
		taskService: service.NewTaskService(storage),
		storage : storage,
	}
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

	tasks, err := h.taskService.GetTasks(statusFilter)
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.taskService.GetTaskByID(id)
	if err != nil {
		http.Error(w, "Failed to get task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
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

	if err := h.taskService.DeleteTask(id, username); err != nil {
		errMsg := err.Error()
	
		// Если юзер не найден — это баг, 500
		if strings.Contains(errMsg, "user not found in storage") {
			log.Printf("Critical: user from session not in DB: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		
		// Если задача не найдена или доступ запрещён — 404
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "access denied") {
			http.Error(w, "Task not found or access denied", http.StatusNotFound)
			return
		}
		
		// Всё остальное — 500
		log.Printf("DeleteTask error: %v", err)
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
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

	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	path = strings.TrimSuffix(path, "/claim")
	taskID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	taskDTO, err :=  h.taskService.ClaimTask(taskID, username)
	if err != nil {
		errMsg := err.Error()

		if strings.Contains(errMsg, "already claimed") {
			http.Error(w, "Task already claimed", http.StatusConflict)
			return
		}

		if strings.Contains(errMsg, "fetch") {
			http.Error(w, "Failed to fetch task", http.StatusInternalServerError)
		}

		http.Error(w, "Failed to claim task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taskDTO)
}

func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	path = strings.TrimSuffix(path, "/complete")
	taskID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	username := r.Context().Value("username").(string)

	task, err := h.taskService.CompleteTask(taskID, username)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "forbidden") {
			http.Error(w, "You can only complete tasks assigned to you", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to complete task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	username := r.Context().Value("username").(string)
	var req model.UpdateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.taskService.UpdateTask(id, req, username)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, "You can't edit this task", http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)

}
