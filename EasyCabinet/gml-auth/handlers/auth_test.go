package handlers

import (
	"bytes"
	"gml-auth/models"
	"gml-auth/storage"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func setupStorage(t *testing.T) *storage.Storage {
	f, _ := os.CreateTemp("", "test-*.json")
	f.WriteString(`{"users":[]}`)
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	s := storage.New(f.Name())
	s.AddUser(models.User{
		UUID: "uuid-1", Login: "GamerVII", Password: "pass123",
		IsSlim: false, Blocked: false,
	})
	s.AddUser(models.User{
		UUID: "uuid-2", Login: "banned", Password: "pass",
		Blocked: true, BlockReason: "читерство",
	})
	return s
}

func TestAuthSuccess(t *testing.T) {
	s := setupStorage(t)
	h := NewAuthHandler(s)
	body := `{"Login":"GamerVII","Password":"pass123","Totp":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/auth/signin", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.SignIn(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "GamerVII") {
		t.Errorf("expected login in response: %s", w.Body.String())
	}
}

func TestAuthWrongPassword(t *testing.T) {
	s := setupStorage(t)
	h := NewAuthHandler(s)
	body := `{"Login":"GamerVII","Password":"wrong","Totp":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/auth/signin", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.SignIn(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthNotFound(t *testing.T) {
	s := setupStorage(t)
	h := NewAuthHandler(s)
	body := `{"Login":"nobody","Password":"pass","Totp":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/auth/signin", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.SignIn(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAuthBlocked(t *testing.T) {
	s := setupStorage(t)
	h := NewAuthHandler(s)
	body := `{"Login":"banned","Password":"pass","Totp":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/auth/signin", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.SignIn(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}
