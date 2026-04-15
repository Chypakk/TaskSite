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
	// "strings"

	"github.com/go-chi/chi/v5"

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

	r := chi.NewRouter()

	r.Use(handler.RequestLogger)

	r.Post("/api/register", userHandler.Register)
	r.Post("/api/login", userHandler.Login)
	r.Post("/api/me", userHandler.GetMe)

	r.Group(func(r chi.Router) {
		r.Use(handler.RequestLogger)
		r.Use(sessionStore.AuthMiddleware)

        r.Post("/api/logout", userHandler.Logout)

		r.Route("/api/tasks", func(r chi.Router) {
			r.Get("/", taskHandler.GetTasks)
			r.Post("/", taskHandler.CreateTask)
			r.Get("/ungrouped", taskHandler.GetUngroupedTasks)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", taskHandler.GetTaskById)
				r.Put("/", taskHandler.UpdateTask)
                r.Delete("/", taskHandler.DeleteTask)
                r.Post("/claim", taskHandler.ClaimTask)
                r.Post("/complete", taskHandler.CompleteTask)
                r.Put("/group", groupHandler.AssignTaskToGroup)
			})
		})

		r.Route("/api/groups", func(r chi.Router) {
            r.Post("/", groupHandler.CreateGroup)
            r.Get("/", groupHandler.GetGroups)
            r.Get("/{id}/tasks", groupHandler.GetGroupTasks)
        })
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        if err := storage.Ping(); err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })
    r.Mount("/swagger/", httpSwagger.WrapHandler)

	fs := http.FileServer(http.Dir("../../static"))
    r.Handle("/*", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
