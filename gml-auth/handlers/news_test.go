package handlers

import (
	"encoding/json"
	"gml-auth/news"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockCache struct {
	items []news.NewsItem
}

func (m *mockCache) Get(limit, offset int) []news.NewsItem {
	if offset >= len(m.items) {
		return []news.NewsItem{}
	}
	end := offset + limit
	if end > len(m.items) {
		end = len(m.items)
	}
	return m.items[offset:end]
}

func TestNewsHandlerDefault(t *testing.T) {
	mock := &mockCache{items: []news.NewsItem{
		{ID: 1, Title: "A"}, {ID: 2, Title: "B"},
	}}
	h := NewNewsHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/news", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var items []news.NewsItem
	json.NewDecoder(w.Body).Decode(&items)
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
}

func TestNewsHandlerLimitOffset(t *testing.T) {
	mock := &mockCache{items: []news.NewsItem{
		{ID: 1}, {ID: 2}, {ID: 3},
	}}
	h := NewNewsHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/news?limit=1&offset=1", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	var items []news.NewsItem
	json.NewDecoder(w.Body).Decode(&items)
	if len(items) != 1 {
		t.Fatalf("expected 1, got %d", len(items))
	}
	if items[0].ID != 2 {
		t.Errorf("expected id 2, got %d", items[0].ID)
	}
}

func TestNewsHandlerEmpty(t *testing.T) {
	mock := &mockCache{}
	h := NewNewsHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/news", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if body != "[]\n" {
		t.Errorf("expected empty array, got: %s", body)
	}
}
