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

	http.HandleFunc("/api/register", userHandler.Register)
	http.HandleFunc("/api/login", userHandler.Login)

	sessionStore := userHandler.GetSessionStore()

	http.HandleFunc("/api/tasks", sessionStore.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			taskHandler.GetTasks(w, r)
		case http.MethodPost:
			taskHandler.CreateTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/tasks/", sessionStore.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/claim") && r.Method == http.MethodPost {
			taskHandler.ClaimTask(w, r)
			return
		}

		if r.Method == http.MethodDelete {
			taskHandler.DeleteTask(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	fs := http.FileServer(http.Dir("../../static"))
	http.Handle("/", fs)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
