package storage

import (
	"gml-auth/models"
	"os"
	"testing"
)

func TestLoadSave(t *testing.T) {
	// временный файл
	f, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"users":[]}`)
	f.Close()
	defer os.Remove(f.Name())

	s := New(f.Name())

	// добавить пользователя
	user := models.User{UUID: "uuid-1", Login: "test", Password: "pass"}
	if err := s.AddUser(user); err != nil {
		t.Fatal(err)
	}

	// найти пользователя
	found, err := s.FindByLogin("test")
	if err != nil {
		t.Fatal(err)
	}
	if found.UUID != "uuid-1" {
		t.Errorf("expected uuid-1, got %s", found.UUID)
	}
}

func TestFindNotFound(t *testing.T) {
	f, _ := os.CreateTemp("", "test-*.json")
	f.WriteString(`{"users":[]}`)
	f.Close()
	defer os.Remove(f.Name())

	s := New(f.Name())
	_, err := s.FindByLogin("nobody")
	if err == nil {
		t.Error("expected error for missing user")
	}
}
