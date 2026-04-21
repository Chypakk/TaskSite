package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"tasksite/internal/model"
	"tasksite/internal/service"
	"tasksite/internal/storage"

	"github.com/go-chi/chi/v5"
)

type TaskGroupHandler struct {
	groupService *service.GroupService
}

func NewTaskGroupHandler(storage *storage.Storage) *TaskGroupHandler {
	return &TaskGroupHandler{groupService: service.NewGroupService(storage)}
}

func (h *TaskGroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req model.CreateGroupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    req.Name = strings.TrimSpace(req.Name)
    if req.Name == "" {
        http.Error(w, "Group name is required", http.StatusBadRequest)
        return
    }
    group, err := h.groupService.CreateTaskGroup(ctx, req.Name, req.Description)
    if err != nil {
        if strings.Contains(err.Error(), "already exists") {
            http.Error(w, "Group already exists", http.StatusConflict)
            return
        }
        http.Error(w, "Failed to create group", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(group)
}

func (h *TaskGroupHandler) GetGroups(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    groups, err := h.groupService.GetTaskGroups(ctx)
    if err != nil {
        http.Error(w, "Failed to get groups", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(groups)
}

func (h *TaskGroupHandler) GetGroupTasks(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    path := strings.TrimPrefix(r.URL.Path, "/api/groups/")
    path = strings.TrimSuffix(path, "/tasks")
    groupID, err := strconv.Atoi(chi.URLParam(r, "id"))
    if err != nil {
        http.Error(w, "Invalid group ID", http.StatusBadRequest)
        return
    }
    
    var statusFilter *string
    if s := r.URL.Query().Get("status"); s != "" {
        statusFilter = &s
    }
    
    tasks, err := h.groupService.GetTasksByGroup(ctx, groupID, statusFilter)
    if err != nil {
        http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tasks)
}

func (h *TaskGroupHandler) AssignTaskToGroup(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
    path = strings.TrimSuffix(path, "/group")
    taskID, err := strconv.Atoi(chi.URLParam(r, "id"))
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }
    
    var req model.AssignTaskToGroupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if err := h.groupService.AssignTaskToGroup(ctx, taskID, req.GroupID); err != nil {
        if strings.Contains(err.Error(), "group not found") {
            http.Error(w, "Group not found", http.StatusNotFound)
            return
        }
        http.Error(w, "Failed to assign task to group", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}