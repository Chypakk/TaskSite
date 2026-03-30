package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"tasksite/internal/model"
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

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Создаёт нового пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  model.RegisterRequest  true  "Данные регистрации"
// @Success      201  {object}  model.User
// @Failure      400  {string}  string  "Invalid request body"
// @Failure      409  {string}  string  "Username already exists"
// @Router       /register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.RegisterRequest

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

// Login godoc
// @Summary      Вход в систему
// @Description  Проверяет логин/пароль и выдаёт токен сессии
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  model.LoginRequest  true  "Данные входа"
// @Success      200  {object}  model.LoginResponse
// @Failure      400  {string}  string  "Invalid request body"
// @Failure      401  {string}  string  "Invalid credentials"
// @Router       /login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.LoginRequest

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
	json.NewEncoder(w).Encode(model.LoginResponse{
		Message:  "Login successful",
		Token:    token,
		Username: user.Username,
		ID:       user.ID,
	})
}

func (h *UserHandler) GetSessionStore() *SessionStore {
	return h.sessionStore
}
