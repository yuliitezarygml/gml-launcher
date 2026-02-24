package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscordFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		messages := []map[string]any{
			{
				"id":        "1234567890123456789",
				"content":   "Обновление сервера\nДобавлены новые плагины и исправлены баги",
				"timestamp": "2024-01-15T12:00:00.000000+00:00",
			},
			{
				"id":        "1234567890123456780",
				"content":   "Краткое объявление",
				"timestamp": "2024-01-14T10:00:00.000000+00:00",
			},
		}
		json.NewEncoder(w).Encode(messages)
	}))
	defer srv.Close()

	p := &DiscordProvider{
		token:   "Bot testtoken",
		channel: "123456789",
		baseURL: srv.URL,
	}

	items, err := p.Fetch(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "Обновление сервера" {
		t.Errorf("unexpected title: %s", items[0].Title)
	}
	if items[0].Description != "Обновление сервера\nДобавлены новые плагины и исправлены баги" {
		t.Errorf("unexpected description: %s", items[0].Description)
	}
}
