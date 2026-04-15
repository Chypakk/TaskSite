package handler

import (
	"context"
	"log"
	"net/http"
	"tasksite/internal/ws"
	"time"

	"github.com/coder/websocket"
)

type WSHandler struct {
	hub          *ws.Hub
	sessionStore *SessionStore
}

func NewWSHandler(hub *ws.Hub, sessionStore *SessionStore) *WSHandler {
	return &WSHandler{
		hub:          hub,
		sessionStore: sessionStore,
	}
}

// ServeWS обрабатывает подключение к вебсокету
func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// 1. Проверяем авторизацию
	// Chi-мидлвар уже положил username в контекст, но для вебсокетов часто токен передают в query-параметре
	// (браузерный WebSocket API не умеет кастомные заголовки)
	token := r.URL.Query().Get("token")
	if token == "" {
		// Фоллбэк на заголовок (для Postman/curl/мобилок)
		token = r.Header.Get("X-Session-Token")
	}

	username, ok := h.sessionStore.ValidateSession(token)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Апгрейд до WebSocket
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Для локалки ок. В проде лучше false + CORS
	})
	if err != nil {
		log.Printf("WS accept failed: %v", err)
		return
	}

	// 3. Создаём контекст с таймаутом (например, 24 часа = время жизни сессии)
	ctx, cancel := context.WithTimeout(r.Context(), 24*time.Hour)
	defer cancel()

	// 4. Создаём клиента и запускаем его
	client := ws.NewClient(h.hub, conn)
	
	log.Printf("WS connected: %s", username)
	client.Start(ctx)
	log.Printf("WS disconnected: %s", username)
}