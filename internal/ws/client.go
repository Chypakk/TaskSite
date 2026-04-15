package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/coder/websocket"
)

// Client представляет одно подключение пользователя.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan Message
	mu   sync.Mutex // Для защиты от конкурентной записи
}

// NewClient создаёт нового клиента.
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan Message, 256),
	}
}

// Start запускает обработку клиента.
func (c *Client) Start(ctx context.Context) {
	c.hub.register <- c
	defer func() {
		c.hub.unregister <- c
	}()

	// Запускаем writer
	go c.writePump(ctx)

	// Читаем сообщения (даже если не используем их, нужно держать коннект)
	for {
		_, _, err := c.conn.Read(ctx)
		if err != nil {
			status := websocket.CloseStatus(err)
			if status != websocket.StatusNormalClosure && status != websocket.StatusGoingAway {
				log.Printf("WebSocket read error: %v", err)
			}
			return
		}
	}
}

// writePump отправляет сообщения клиенту.
func (c *Client) writePump(ctx context.Context) {
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}

			payload, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			c.mu.Lock()
			err = c.conn.Write(ctx, websocket.MessageText, payload)
			c.mu.Unlock()

			if err != nil {
				log.Printf("Error writing to websocket: %v", err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}
