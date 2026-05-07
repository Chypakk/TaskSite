package handler

import (
	"errors"
	"net/http"
	internalErrors "tasksite/internal/errors"
)

func RespondError(w http.ResponseWriter, err error) {
	switch{
	case errors.Is(err, internalErrors.ErrNotFound):
		http.Error(w, "Not found", http.StatusNotFound)
	case errors.Is(err, internalErrors.ErrAlreadyExists):
        http.Error(w, "Already exists", http.StatusConflict)
    case errors.Is(err, internalErrors.ErrAccessDenied):
        http.Error(w, "Access denied", http.StatusForbidden)
    case errors.Is(err, internalErrors.ErrInvalidInput):
        http.Error(w, "Invalid input", http.StatusBadRequest)
    default:
        http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}