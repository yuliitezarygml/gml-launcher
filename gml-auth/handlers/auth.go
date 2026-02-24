package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"gml-auth/models"
	"gml-auth/storage"
	"net/http"
)

type AuthHandler struct {
	store *storage.Storage
}

func NewAuthHandler(store *storage.Storage) *AuthHandler {
	return &AuthHandler{store: store}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Message: "Неверный формат запроса"})
		return
	}

	user, err := h.store.FindByLogin(req.Login)
	if errors.Is(err, storage.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Message: "Пользователь не найден"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Message: "Ошибка сервера"})
		return
	}

	if user.Blocked {
		msg := fmt.Sprintf("Пользователь заблокирован. Причина: %s", user.BlockReason)
		writeJSON(w, http.StatusForbidden, models.ErrorResponse{Message: msg})
		return
	}

	if user.Password != req.Password {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Message: "Неверный логин или пароль"})
		return
	}

	writeJSON(w, http.StatusOK, models.AuthResponse{
		Login:    user.Login,
		UserUuid: user.UUID,
		IsSlim:   user.IsSlim,
		Message:  "Успешная авторизация",
	})
}

// Refresh — GET /api/v1/users/refresh
// Сервер не хранит сессии, поэтому всегда 401.
// GML Launcher web-панель при 401 перенаправит на страницу входа.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusUnauthorized, models.WebErrorResponse{
		Errors: []string{"Сессия не найдена. Выполните вход."},
	})
}
