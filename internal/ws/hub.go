package ws

import (
	"context"
	"sync"
)

type Hub struct {
	// Соединения: ключ - указатель на клиента
	clients map[*Client]struct{}

	// Каналы для управления состоянием
	broadcast  chan Message // Сообщения для рассылки
	register   chan *Client // Регистрация новых клиентов
	unregister chan *Client // Удаление клиентов

	mu sync.RWMutex // Мьютекс для безопасности работы с map clients
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			h.mu.Unlock()
		
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients{
				select{
				case client.send <- message:

				default:
				}
			}
			h.mu.RUnlock()

		case <-ctx.Done():
			h.mu.Lock()
			for client := range h.clients{
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			return
		}
	}
}

func (h *Hub) Broadcast(msg Message) {
	select {
	case h.broadcast <- msg:
	default:
	}
}