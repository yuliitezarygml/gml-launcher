package storage

import (
	"encoding/json"
	"errors"
	"gml-auth/models"
	"os"
	"sync"
)

var ErrNotFound = errors.New("user not found")

type Storage struct {
	mu       sync.RWMutex
	filePath string
}

func New(filePath string) *Storage {
	return &Storage{filePath: filePath}
}

func (s *Storage) load() (models.Database, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return models.Database{}, err
	}
	var db models.Database
	return db, json.Unmarshal(data, &db)
}

func (s *Storage) save(db models.Database) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Storage) FindByLogin(login string) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	db, err := s.load()
	if err != nil {
		return models.User{}, err
	}
	for _, u := range db.Users {
		if u.Login == login {
			return u, nil
		}
	}
	return models.User{}, ErrNotFound
}

func (s *Storage) ListUsers() ([]models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	db, err := s.load()
	if err != nil {
		return nil, err
	}
	return db.Users, nil
}

func (s *Storage) AddUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, err := s.load()
	if err != nil {
		return err
	}
	db.Users = append(db.Users, user)
	return s.save(db)
}

func (s *Storage) DeleteUser(login string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, err := s.load()
	if err != nil {
		return err
	}
	filtered := db.Users[:0]
	for _, u := range db.Users {
		if u.Login != login {
			filtered = append(filtered, u)
		}
	}
	if len(filtered) == len(db.Users) {
		return ErrNotFound
	}
	db.Users = filtered
	return s.save(db)
}

func (s *Storage) UpdateUser(login string, fn func(*models.User)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, err := s.load()
	if err != nil {
		return err
	}
	found := false
	for i := range db.Users {
		if db.Users[i].Login == login {
			fn(&db.Users[i])
			found = true
			break
		}
	}
	if !found {
		return ErrNotFound
	}
	return s.save(db)
}
