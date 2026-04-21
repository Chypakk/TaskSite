package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"tasksite/internal/logger"
	"tasksite/internal/model"
	"tasksite/internal/service"
	"tasksite/internal/storage"
	"tasksite/internal/ws"

	"github.com/go-chi/chi/v5"
)

type TaskHandler struct {
	taskService *service.TaskService
	wsHub       *ws.Hub
}

func NewTaskHandler(storage *storage.Storage, wsHub *ws.Hub) *TaskHandler {
	return &TaskHandler{
		taskService: service.NewTaskService(storage, storage, storage),
		wsHub: wsHub,
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
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var req model.CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(ctx, "CreateTask: decode failed", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		log.Info(ctx, "CreateTask: validation failed", "reason", "empty name")
		http.Error(w, "Task name is required", http.StatusBadRequest)
		return
	}

	task, err := h.taskService.CreateTask(ctx, req.Name, req.Description, req.Author)
	if err != nil {
		log.Error(ctx, "CreateTask: storage error", err, "name", req.Name)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	sendEvent(h.wsHub, ws.EventTaskCreated, task)

	log.Info(ctx, "CreateTask: success", "task_id", task.ID)

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
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.URL.Query().Get("page") != "" || r.URL.Query().Get("limit") != "" {
        pq := parsePagination(r)
        
        var groupID *int
        if groupParam := chi.URLParam(r, "id"); groupParam != "" {
            if gid, err := strconv.Atoi(groupParam); err == nil {
                groupID = &gid
            }
        }
        
        resp, err := h.taskService.GetTasksPaginated(ctx, pq, groupID)
        if err != nil {
            log.Error(ctx, "GetTasks: service error", err, "pagination", pq)
            http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
        return
    }

	var statusFilter *string

	if status := r.URL.Query().Get("status"); status != "" {
		statusFilter = &status
	}

	tasks, err := h.taskService.GetTasks(ctx, statusFilter)
	if err != nil {
		log.Error(ctx, "GetTasks: service error", err, "status_filter", statusFilter)
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.Atoi(path)
	if err != nil {
		log.Info(ctx, "invalid id", "taskID", path)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.taskService.GetTaskByID(ctx, id)
	if err != nil {
		log.Error(ctx, "GetTaskById: service error", err, "task_id", id)
		http.Error(w, "Failed to get task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Info(ctx, "invalid id", "taskID", path)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	username, ok := r.Context().Value("username").(string)
	if !ok {
		log.Warn(ctx, "username not in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.taskService.DeleteTask(ctx, id, username); err != nil {
		errMsg := err.Error()

		// Если юзер не найден — это баг, 500
		if strings.Contains(errMsg, "user not found in storage") {
			log.Error(ctx, "DeleteTask: CRITICAL - user in session not in DB", err, "username", username)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Если задача не найдена или доступ запрещён — 404
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "access denied") {
			log.Info(ctx, "not found or denied", "taskId", id, "username", username)
			http.Error(w, "Task not found or access denied", http.StatusNotFound)
			return
		}

		// Всё остальное — 500
		log.Error(ctx, "DeleteTask: unexpected error", err, "task_id", id, "username", username)
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	sendEvent(h.wsHub, ws.EventTaskDeleted, map[string]int {"id": id})

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) ClaimTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	username, ok := r.Context().Value("username").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	path = strings.TrimSuffix(path, "/claim")

	taskID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Info(ctx, "invalid id", "taskID", path)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	taskDTO, err := h.taskService.ClaimTask(ctx, taskID, username)
	if err != nil {
		errMsg := err.Error()

		if strings.Contains(errMsg, "already claimed") {
			log.Info(ctx, "already claimed", "taskID", taskID)
			http.Error(w, "Task already claimed", http.StatusConflict)
			return
		}

		if strings.Contains(errMsg, "fetch") {
			log.Error(ctx, "ClaimTask: failed to fetch task", err, "task_id", taskID)
			http.Error(w, "Failed to fetch task", http.StatusInternalServerError)
			return
		}

		log.Error(ctx, "ClaimTask: error", err, "task_id", taskID, "username", username)
		http.Error(w, "Failed to claim task", http.StatusInternalServerError)
		return
	}

	sendEvent(h.wsHub, ws.EventTaskClaimed, taskDTO)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taskDTO)
}

func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	path = strings.TrimSuffix(path, "/complete")
	taskID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Info(ctx, "invalid id", "taskID", path)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	username := r.Context().Value("username").(string)

	// task, err := h.taskService.CompleteTask(ctx, taskID, username)
	err = h.taskService.CompleteTask(ctx, taskID, username)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Warn(ctx, "Task not found ", "task_id", taskID)
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "forbidden") {
			log.Info(ctx, "Access denied", "taskID", taskID, "username", username)
			http.Error(w, "You can only complete tasks assigned to you", http.StatusForbidden)
			return
		}
		log.Error(ctx, "Failed to complete task", err, "taskID", taskID)
		http.Error(w, "Failed to complete task", http.StatusInternalServerError)
		return
	}

	sendEvent(h.wsHub, ws.EventTaskCompleted, nil)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Info(ctx, "invalid id", "taskID", path)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	username := r.Context().Value("username").(string)
	var req model.UpdateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn(ctx, "Invalid request body", "err", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.taskService.UpdateTask(ctx, id, req, username)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") {
			log.Info(ctx, "access denied", "taskId", id, "username", username)
			http.Error(w, "You can't edit this task", http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			log.Warn(ctx, "Task not found", "task_id", id, "err_msg", err.Error())
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		log.Error(ctx, "Failed to update task", err, "taskID", id)
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	sendEvent(h.wsHub, ws.EventTaskUpdated, task)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)

}

func (h *TaskHandler) GetUngroupedTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var statusFilter *string
	if s := r.URL.Query().Get("status"); s != "" {
		statusFilter = &s
	}

	tasks, err := h.taskService.GetUngroupedTasks(ctx, statusFilter)
	if err != nil {
		log.Error(ctx, "GetUngroupedTasks: service error", err)
		http.Error(w, "Failed to get ungrouped tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func parsePagination(r *http.Request) model.PaginationQuery {
    pq := model.DefaultPagination()
    
    if page := r.URL.Query().Get("page"); page != "" {
        if p, err := strconv.Atoi(page); err == nil {
            pq.Page = p
        }
    }
    if limit := r.URL.Query().Get("limit"); limit != "" {
        if l, err := strconv.Atoi(limit); err == nil {
            pq.Limit = l
        }
    }
    pq.Status = r.URL.Query().Get("status")
    pq.Sort = r.URL.Query().Get("sort")
    
    pq.Validate()
    return pq
}