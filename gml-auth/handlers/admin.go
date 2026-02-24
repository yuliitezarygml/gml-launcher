package handlers

import (
	"encoding/json"
	"errors"
	"gml-auth/models"
	"gml-auth/storage"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type AdminHandler struct {
	store *storage.Storage
}

func NewAdminHandler(store *storage.Storage) *AdminHandler {
	return &AdminHandler{store: store}
}

// loginFromPath извлекает {login} из пути вида /admin/users/{login}[/action]
func loginFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// /admin/users/{login} → ["admin","users","{login}"]
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path // /admin/users, /admin/users/foo, /admin/users/foo/block

	switch {
	case path == "/admin/users" && r.Method == http.MethodGet:
		h.listUsers(w, r)
	case path == "/admin/users" && r.Method == http.MethodPost:
		h.createUser(w, r)
	case strings.HasSuffix(path, "/block") && r.Method == http.MethodPatch:
		login := loginFromPath(strings.TrimSuffix(path, "/block"))
		h.blockUser(w, r, login)
	case strings.HasSuffix(path, "/unblock") && r.Method == http.MethodPatch:
		login := loginFromPath(strings.TrimSuffix(path, "/unblock"))
		h.unblockUser(w, r, login)
	case r.Method == http.MethodDelete:
		login := loginFromPath(path)
		h.deleteUser(w, r, login)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *AdminHandler) listUsers(w http.ResponseWriter, _ *http.Request) {
	users, err := h.store.ListUsers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Message: "Ошибка чтения"})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *AdminHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Login == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Message: "Нужны login и password"})
		return
	}
	user := models.User{
		UUID:     uuid.New().String(),
		Login:    req.Login,
		Password: req.Password,
		IsSlim:   req.IsSlim,
	}
	if err := h.store.AddUser(user); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Message: "Ошибка создания"})
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (h *AdminHandler) deleteUser(w http.ResponseWriter, _ *http.Request, login string) {
	if login == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Message: "Нужен login"})
		return
	}
	err := h.store.DeleteUser(login)
	if errors.Is(err, storage.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Message: "Пользователь не найден"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) blockUser(w http.ResponseWriter, r *http.Request, login string) {
	if login == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Message: "Нужен login"})
		return
	}
	var req models.BlockRequest
	json.NewDecoder(r.Body).Decode(&req)

	err := h.store.UpdateUser(login, func(u *models.User) {
		u.Blocked = true
		u.BlockReason = req.Reason
	})
	if errors.Is(err, storage.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Message: "Пользователь не найден"})
		return
	}
	writeJSON(w, http.StatusOK, models.ErrorResponse{Message: "Заблокирован"})
}

func (h *AdminHandler) unblockUser(w http.ResponseWriter, _ *http.Request, login string) {
	if login == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Message: "Нужен login"})
		return
	}
	err := h.store.UpdateUser(login, func(u *models.User) {
		u.Blocked = false
		u.BlockReason = ""
	})
	if errors.Is(err, storage.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Message: "Пользователь не найден"})
		return
	}
	writeJSON(w, http.StatusOK, models.ErrorResponse{Message: "Разблокирован"})
}
