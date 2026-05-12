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
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tasksite"
	"time"

	"github.com/go-chi/chi/v5"

	"tasksite/internal/config"
	"tasksite/internal/handler"
	"tasksite/internal/storage"
	"tasksite/internal/ws"

	_ "tasksite/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)


func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	// заглушка, в дальнейшем это будет основной метод
	storage.NewConnectDB(cfg)

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "tasksite.db"
	}
	storage, err := storage.ConnectDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	wsHub := ws.NewHub()
	go wsHub.Run(ctx)

	taskHandler := handler.NewTaskHandler(storage, wsHub)
	userHandler := handler.NewUserHandler(storage)
	groupHandler := handler.NewTaskGroupHandler(storage)

	sessionStore := userHandler.GetSessionStore()

	wsHandler := handler.NewWSHandler(wsHub, sessionStore)

	r := chi.NewRouter()

	r.Use(handler.RequestLogger)

	r.Post("/api/register", userHandler.Register)
	r.Post("/api/login", userHandler.Login)
	r.Post("/api/me", userHandler.GetMe)

	r.Group(func(r chi.Router) {
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
			r.Put("/{id}", groupHandler.EditGroup)
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

	// fs := http.FileServer(http.Dir("../../static"))
	// r.Handle("/*", fs)

	contentStatic, _ := fs.Sub(tasksite.StaticFS, "static")
    fileServer := http.FileServer(http.FS(contentStatic))
    
	// r.Route("/static", func(r chi.Router) {
	// 	r.Handle("/*", http.StripPrefix("/static/", fileServer))
	// })

	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))
	
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        index, err := contentStatic.Open("index.html")
        if err != nil {
            http.Error(w, "Index not found", http.StatusNotFound)
            return
        }
        defer index.Close()

        http.ServeContent(w, r, "index.html", time.Now(), index.(io.ReadSeeker))
    })

	mux := http.NewServeMux()

	mux.HandleFunc("/api/ws", wsHandler.ServeWS)
	mux.Handle("/", r)

	// Запускаем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Server starting on port %s", port)
		serverErrors <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	case <-stop:
		log.Println("Shutdown signal received, starting graceful shutdown...")
	}

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if err := storage.Close(); err != nil {
		log.Printf("Database close error: %v", err)
	}

	log.Println("Application stopped gracefully")

	// log.Printf("Server starting on port %s", port)
	// if err := http.ListenAndServe(":"+port, mux); err != nil {
	//     log.Fatalf("Server failed: %v", err)
	// }
}
