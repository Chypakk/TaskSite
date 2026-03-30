package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"tasksite/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	storage      *storage.Storage
	sessionStore *SessionStore
}

func NewUserHandler(storage *storage.Storage) *UserHandler {
	return &UserHandler{
		storage:      storage,
		sessionStore: NewSessionStore(),
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.storage.CreateUser(req.Username, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Username already exists", http.StatusConflict) // 409
			return
		}

		http.Error(w, "Failed to register", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := h.sessionStore.CreateSession(user.Username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Login successful",
		"token":    token,
		"username": user.Username,
		"id":       user.ID,
	})
}

func (h *UserHandler) GetSessionStore() *SessionStore {
	return h.sessionStore
}
