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

func NewTaskHandler(storage *storage.Storage) *TaskHandler{
	return &TaskHandler{storage: storage}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct{
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, "Task name is required", http.StatusBadRequest)
		return
	}

	task, err := h.storage.CreateTask(req.Name)
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request)  {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks, err := h.storage.GetTasks()
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

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request)  {
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

	if err := h.storage.DeleteTask(id); err != nil {
		http.Error(w, "Failed to delete task", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}