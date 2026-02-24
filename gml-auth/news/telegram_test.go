package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTelegramFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"ok": true,
			"result": []map[string]any{
				{
					"update_id": 100,
					"channel_post": map[string]any{
						"message_id": 42,
						"chat":       map[string]any{"username": "sinkdev_dev"},
						"text":       "Заголовок новости\nТекст новости подробнее",
						"date":       1700000000,
					},
				},
				{
					"update_id": 101,
					"channel_post": map[string]any{
						"message_id": 43,
						"chat":       map[string]any{"username": "sinkdev_dev"},
						"text":       "Одна строка",
						"date":       1700001000,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := &TelegramProvider{
		token:   "testtoken",
		channel: "sinkdev_dev",
		baseURL: srv.URL,
	}

	items, err := p.Fetch(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	// Newest first (date 1700001000 > 1700000000)
	if items[0].ID != 43 {
		t.Errorf("expected id 43 first, got %d", items[0].ID)
	}
	if items[0].Title != "Одна строка" {
		t.Errorf("unexpected title: %s", items[0].Title)
	}
	if items[1].ID != 42 {
		t.Errorf("expected id 42 second, got %d", items[1].ID)
	}
	if items[1].Title != "Заголовок новости" {
		t.Errorf("unexpected title: %s", items[1].Title)
	}
}

func TestTelegramFetchFiltersOtherChannels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"ok": true,
			"result": []map[string]any{
				{
					"update_id": 200,
					"channel_post": map[string]any{
						"message_id": 10,
						"chat":       map[string]any{"username": "other_channel"},
						"text":       "Чужой пост",
						"date":       1700000000,
					},
				},
				{
					"update_id": 201,
					"channel_post": map[string]any{
						"message_id": 11,
						"chat":       map[string]any{"username": "sinkdev_dev"},
						"text":       "Наш пост",
						"date":       1700001000,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := &TelegramProvider{
		token:   "testtoken",
		channel: "sinkdev_dev",
		baseURL: srv.URL,
	}

	items, err := p.Fetch(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item (filtered), got %d", len(items))
	}
	if items[0].ID != 11 {
		t.Errorf("expected id 11, got %d", items[0].ID)
	}
}

func TestTelegramFetchOffsetAdvances(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var result []map[string]any
		if callCount == 1 {
			result = []map[string]any{
				{
					"update_id": 300,
					"channel_post": map[string]any{
						"message_id": 20,
						"chat":       map[string]any{"username": "testchan"},
						"text":       "Пост 1",
						"date":       1700000000,
					},
				},
			}
		}
		// Second call returns empty (offset advanced past 300)
		resp := map[string]any{"ok": true, "result": result}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := &TelegramProvider{
		token:   "testtoken",
		channel: "testchan",
		baseURL: srv.URL,
	}

	items1, _ := p.Fetch(10)
	if len(items1) != 1 {
		t.Fatalf("first fetch: expected 1, got %d", len(items1))
	}
	if p.offset != 301 {
		t.Errorf("expected offset 301, got %d", p.offset)
	}

	items2, _ := p.Fetch(10)
	// Stored items are still returned (accumulation)
	if len(items2) != 1 {
		t.Fatalf("second fetch: expected 1 stored item, got %d", len(items2))
	}
}
