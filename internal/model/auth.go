package model

// RegisterRequest запрос регистрации
type RegisterRequest struct {
	Username string `json:"username" example:"test"`
	Password string `json:"password" example:"secret123"`
}

// LoginRequest запрос входа
type LoginRequest struct {
	Username string `json:"username" example:"test"`
	Password string `json:"password" example:"secret123"`
}

// LoginResponse ответ при входе
type LoginResponse struct {
	Message  string `json:"message" example:"Login successful"`
	Token    string `json:"token" example:"a1b2c3d4e5f6..."`
	Username string `json:"username" example:"test"`
	ID       int    `json:"id" example:"1"`
}

// CreateTaskRequest запрос создания задачи
type CreateTaskRequest struct {
	Name        string `json:"name" example:"Задача 1"`
	Description string `json:"description" example:"Описание"`
	Author      string `json:"author" example:"Пользователь1"`
}

type UpdateTaskRequest struct {
	Name        string `json:"name,omitempty" example:"Задача 1"`
	Description string `json:"description,omitempty" example:"Описание"`
	Author      string `json:"author,omitempty" example:"Пользователь1"`
	Status      string `json:"status,omitempty" example:"open"`
}
