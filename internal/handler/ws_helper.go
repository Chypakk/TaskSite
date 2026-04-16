package handler

import (
	"log"
	"tasksite/internal/ws"
)

// sendEvent — хелпер для отправки события в хаб
func sendEvent(hub *ws.Hub, eventType string, payload any) {
	msg, err := ws.NewMessage(eventType, payload)
	if err != nil {
		log.Printf("Failed to create WS message: %v", err)
		return
	}
	hub.Broadcast(msg)
}