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
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"

	"tasksite/internal/handler"
	"tasksite/internal/storage"
	"tasksite/internal/ws"

	_ "tasksite/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

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

	// wsHandler := handler.NewWSHandler(wsHub, sessionStore)

	r := chi.NewRouter()

	r.Use(handler.RequestLogger)

	r.Post("/api/register", userHandler.Register)
	r.Post("/api/login", userHandler.Login)
	r.Post("/api/me", userHandler.GetMe)
	
	// r.Get("/api/ws", wsHandler.ServeWS)

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

	mux := http.NewServeMux()

    // 1. WebSocket роут — ПРЯМОЙ, без Chi middleware!
    mux.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
        // Авторизация вручную
        token := r.URL.Query().Get("token")
		ctx := r.Context()
        if token == "" {
            token = r.Header.Get("X-Session-Token")
        }
        username, ok := sessionStore.ValidateSession(ctx, token)
        if !ok {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
            InsecureSkipVerify: true,
        })
        if err != nil {
            log.Printf("WS accept failed: %v", err)
            return
        }

        ctx, cancel := context.WithTimeout(r.Context(), 24*time.Hour)
        defer cancel()

        client := ws.NewClient(wsHub, conn)
        log.Printf("WS connected: %s", username)
        client.Start(ctx)
        log.Printf("WS disconnected: %s", username)
    })

    // 2. Все остальные роуты — через Chi
    mux.Handle("/", r)

    // Запускаем сервер
    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }

	srv := &http.Server{
        Addr:    ":" + port,
        Handler: mux,
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
