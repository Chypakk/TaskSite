// @title        TaskSite API
// @version      1.0
// @description  Простая тикет-система с авторизацией
// @host         localhost:8080
// @BasePath     /api
// @securityDefinitions.apikey  SessionToken
// @in           header
// @name         X-Session-Token
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"tasksite/internal/handler"
	"tasksite/internal/storage"

	_ "tasksite/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "tasksite.db"
	}
	storage, err := storage.ConnectDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	taskHandler := handler.NewTaskHandler(storage)
	userHandler := handler.NewUserHandler(storage)
	groupHandler := handler.NewTaskGroupHandler(storage)

	sessionStore := userHandler.GetSessionStore()

	http.HandleFunc("/api/register", handler.RequestLogger(userHandler.Register))
	http.HandleFunc("/api/login", handler.RequestLogger(userHandler.Login))
	http.HandleFunc("/api/me", handler.RequestLogger(userHandler.GetMe))
	http.HandleFunc("/api/logout", sessionStore.AuthMiddleware(handler.RequestLogger(userHandler.Logout)))

	http.HandleFunc("/api/tasks", sessionStore.AuthMiddleware(handler.RequestLogger(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				taskHandler.GetTasks(w, r)
			case http.MethodPost:
				taskHandler.CreateTask(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	http.HandleFunc("/api/tasks/", sessionStore.AuthMiddleware(
		handler.RequestLogger(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if strings.HasSuffix(r.URL.Path, "/group") && r.Method == http.MethodPut {
				groupHandler.AssignTaskToGroup(w, r)
				return
			}
			if strings.HasSuffix(path, "/claim") && r.Method == http.MethodPost {
				taskHandler.ClaimTask(w, r)
				return
			}

			if strings.HasSuffix(path, "/complete") && r.Method == http.MethodPost {
				taskHandler.CompleteTask(w, r)
				return
			}

			switch r.Method {
			case http.MethodGet:
				taskHandler.GetTaskById(w, r)
				return
			case http.MethodDelete:
				taskHandler.DeleteTask(w, r)
				return
			case http.MethodPut:
				taskHandler.UpdateTask(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
		})))

	http.HandleFunc("/api/tasks/ungrouped", sessionStore.AuthMiddleware(
		handler.RequestLogger(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			taskHandler.GetUngroupedTasks(w, r)
		})))

	http.HandleFunc("/api/groups", sessionStore.AuthMiddleware(handler.RequestLogger(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				groupHandler.CreateGroup(w, r)
			case http.MethodGet:
				groupHandler.GetGroups(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))

	http.HandleFunc("/api/groups/", sessionStore.AuthMiddleware(handler.RequestLogger(
		func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/tasks") && r.Method == http.MethodGet {
				groupHandler.GetGroupTasks(w, r)
				return
			}
			http.Error(w, "Not found", http.StatusNotFound)
		})))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/health", handler.RequestLogger(
		func(w http.ResponseWriter, r *http.Request) {
			if err := storage.Ping(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))

	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	fs := http.FileServer(http.Dir("../../static"))
	http.Handle("/", fs)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
