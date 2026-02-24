# News Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `GET /api/news` endpoint that serves news from Telegram or Discord channel, configured via `config.json`.

**Architecture:** In-memory cache filled by background goroutine every 60s from Telegram/Discord API. Handler serves from cache instantly. Source selected by `config.json` field `news.source`.

**Tech Stack:** Go 1.21+, stdlib only for HTTP. Telegram Bot API (HTTP). Discord API (HTTP). No new external packages — `github.com/google/uuid` already in go.mod.

---

### Task 1: Config loader

**Files:**
- Create: `config/config.go`
- Create: `config/config_test.go`

**Step 1: Написать тест**

Создать `config/config_test.go`:
```go
package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	f, _ := os.CreateTemp("", "cfg-*.json")
	f.WriteString(`{"news":{"source":"telegram","token":"tok123","channel":"@test","refresh_seconds":30}}`)
	f.Close()
	defer os.Remove(f.Name())

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.News.Source != "telegram" {
		t.Errorf("expected telegram, got %s", cfg.News.Source)
	}
	if cfg.News.Token != "tok123" {
		t.Errorf("expected tok123, got %s", cfg.News.Token)
	}
	if cfg.News.RefreshSeconds != 30 {
		t.Errorf("expected 30, got %d", cfg.News.RefreshSeconds)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	f, _ := os.CreateTemp("", "cfg-*.json")
	f.WriteString(`{"news":{"source":"discord","token":"tok","channel":"123"}}`)
	f.Close()
	defer os.Remove(f.Name())

	cfg, _ := Load(f.Name())
	if cfg.News.RefreshSeconds != 60 {
		t.Errorf("expected default 60, got %d", cfg.News.RefreshSeconds)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	cfg, err := Load("/nonexistent/config.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
	_ = cfg
}
```

**Step 2: Запустить тест — должен упасть**

```bash
go test ./config/...
```

Expected: FAIL — `Load` не определён

**Step 3: Реализовать**

Создать `config/config.go`:
```go
package config

import (
	"encoding/json"
	"os"
)

type NewsConfig struct {
	Source         string `json:"source"`          // "telegram" или "discord"
	Token          string `json:"token"`            // Bot token
	Channel        string `json:"channel"`          // @username или channel ID
	RefreshSeconds int    `json:"refresh_seconds"`  // default 60
}

type Config struct {
	News NewsConfig `json:"news"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.News.RefreshSeconds == 0 {
		cfg.News.RefreshSeconds = 60
	}
	return cfg, nil
}
```

**Step 4: Запустить тест — должен пройти**

```bash
go test ./config/...
```

Expected: PASS

**Step 5: Commit**

```bash
git add config/
git commit -m "feat: add config loader"
```

---

### Task 2: NewsItem модель и интерфейс провайдера

**Files:**
- Create: `news/provider.go`
- Create: `news/provider_test.go`

**Step 1: Написать тест**

Создать `news/provider_test.go`:
```go
package news

import (
	"testing"
	"time"
)

func TestNewsItemFields(t *testing.T) {
	item := NewsItem{
		ID:          1,
		Title:       "Test",
		Description: "Body",
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if item.ID != 1 {
		t.Error("ID mismatch")
	}
	if item.Title != "Test" {
		t.Error("Title mismatch")
	}
}
```

**Step 2: Запустить тест — должен упасть**

```bash
go test ./news/...
```

Expected: FAIL — `NewsItem` не определён

**Step 3: Реализовать**

Создать `news/provider.go`:
```go
package news

// NewsItem — одна новость в формате GML Launcher
type NewsItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
}

// Provider — источник новостей
type Provider interface {
	Fetch(limit int) ([]NewsItem, error)
}
```

**Step 4: Запустить тест — должен пройти**

```bash
go test ./news/...
```

Expected: PASS

**Step 5: Commit**

```bash
git add news/provider.go news/provider_test.go
git commit -m "feat: add NewsItem model and Provider interface"
```

---

### Task 3: Telegram провайдер

**Files:**
- Create: `news/telegram.go`
- Create: `news/telegram_test.go`

**Step 1: Написать тест**

Создать `news/telegram_test.go`:
```go
package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTelegramFetch(t *testing.T) {
	// Мок Telegram Bot API
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"ok": true,
			"result": []map[string]any{
				{
					"message_id": 42,
					"text":       "Заголовок новости\nТекст новости подробнее",
					"date":       1700000000,
				},
				{
					"message_id": 41,
					"text":       "Одна строка",
					"date":       1699999000,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := &TelegramProvider{
		token:   "testtoken",
		channel: "@test",
		baseURL: srv.URL,
	}

	items, err := p.Fetch(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != 42 {
		t.Errorf("expected id 42, got %d", items[0].ID)
	}
	if items[0].Title != "Заголовок новости" {
		t.Errorf("unexpected title: %s", items[0].Title)
	}
	if items[1].Title != "Одна строка" {
		t.Errorf("unexpected title for single-line: %s", items[1].Title)
	}
}
```

**Step 2: Запустить тест — должен упасть**

```bash
go test ./news/... -run TestTelegramFetch
```

Expected: FAIL

**Step 3: Реализовать**

Создать `news/telegram.go`:
```go
package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const telegramBaseURL = "https://api.telegram.org"

type TelegramProvider struct {
	token   string
	channel string
	baseURL string // переопределяется в тестах
}

func NewTelegramProvider(token, channel string) *TelegramProvider {
	return &TelegramProvider{token: token, channel: channel, baseURL: telegramBaseURL}
}

type tgResponse struct {
	OK     bool        `json:"ok"`
	Result []tgMessage `json:"result"`
}

type tgMessage struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"text"`
	Date      int64  `json:"date"`
}

func (p *TelegramProvider) Fetch(limit int) ([]NewsItem, error) {
	url := fmt.Sprintf("%s/bot%s/getUpdates?chat_id=%s&limit=%d",
		p.baseURL, p.token, p.channel, limit)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tgResp tgResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, err
	}

	items := make([]NewsItem, 0, len(tgResp.Result))
	for _, msg := range tgResp.Result {
		if msg.Text == "" {
			continue
		}
		title, description := splitTitle(msg.Text)
		items = append(items, NewsItem{
			ID:          msg.MessageID,
			Title:       title,
			Description: msg.Text,
			CreatedAt:   time.Unix(msg.Date, 0).UTC().Format(time.RFC3339),
		})
		_ = description
	}
	return items, nil
}

// splitTitle возвращает первую строку как заголовок
func splitTitle(text string) (title, rest string) {
	parts := strings.SplitN(text, "\n", 2)
	title = parts[0]
	if len(title) > 100 {
		title = title[:100]
	}
	if len(parts) > 1 {
		rest = parts[1]
	}
	return
}
```

**Step 4: Запустить тест — должен пройти**

```bash
go test ./news/... -run TestTelegramFetch
```

Expected: PASS

**Step 5: Commit**

```bash
git add news/telegram.go news/telegram_test.go
git commit -m "feat: add Telegram news provider"
```

---

### Task 4: Discord провайдер

**Files:**
- Create: `news/discord.go`
- Create: `news/discord_test.go`

**Step 1: Написать тест**

Создать `news/discord_test.go`:
```go
package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscordFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем Authorization header
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
```

**Step 2: Запустить тест — должен упасть**

```bash
go test ./news/... -run TestDiscordFetch
```

Expected: FAIL

**Step 3: Реализовать**

Создать `news/discord.go`:
```go
package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const discordBaseURL = "https://discord.com/api/v10"

type DiscordProvider struct {
	token   string
	channel string
	baseURL string
}

func NewDiscordProvider(token, channel string) *DiscordProvider {
	return &DiscordProvider{
		token:   "Bot " + token,
		channel: channel,
		baseURL: discordBaseURL,
	}
}

type discordMessage struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

func (p *DiscordProvider) Fetch(limit int) ([]NewsItem, error) {
	url := fmt.Sprintf("%s/channels/%s/messages?limit=%d", p.baseURL, p.channel, limit)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", p.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var messages []discordMessage
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, err
	}

	items := make([]NewsItem, 0, len(messages))
	for _, msg := range messages {
		if msg.Content == "" {
			continue
		}
		title, _ := splitTitle(msg.Content)
		// Discord snowflake ID — берём последние 9 цифр как int
		numID := snowflakeToInt(msg.ID)
		items = append(items, NewsItem{
			ID:          numID,
			Title:       title,
			Description: msg.Content,
			CreatedAt:   normalizeTimestamp(msg.Timestamp),
		})
	}
	return items, nil
}

func snowflakeToInt(id string) int {
	if len(id) > 9 {
		id = id[len(id)-9:]
	}
	n, _ := strconv.Atoi(id)
	return n
}

// normalizeTimestamp обрезает микросекунды и offset для RFC3339
func normalizeTimestamp(ts string) string {
	// "2024-01-15T12:00:00.000000+00:00" → "2024-01-15T12:00:00Z"
	if len(ts) >= 19 {
		return ts[:19] + "Z"
	}
	return ts
}
```

**Step 4: Запустить тест — должен пройти**

```bash
go test ./news/... -run TestDiscordFetch
```

Expected: PASS

**Step 5: Commit**

```bash
git add news/discord.go news/discord_test.go
git commit -m "feat: add Discord news provider"
```

---

### Task 5: Кэш с фоновым обновлением

**Files:**
- Create: `news/cache.go`
- Create: `news/cache_test.go`

**Step 1: Написать тест**

Создать `news/cache_test.go`:
```go
package news

import (
	"errors"
	"testing"
	"time"
)

type mockProvider struct {
	items []NewsItem
	err   error
	calls int
}

func (m *mockProvider) Fetch(limit int) ([]NewsItem, error) {
	m.calls++
	return m.items, m.err
}

func TestCacheServesFreshData(t *testing.T) {
	mock := &mockProvider{items: []NewsItem{{ID: 1, Title: "Test"}}}
	c := NewCache(mock, 100*time.Millisecond)
	c.Start()
	defer c.Stop()

	time.Sleep(20 * time.Millisecond)
	items := c.Get(10, 0)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != 1 {
		t.Errorf("expected id 1, got %d", items[0].ID)
	}
}

func TestCacheServesStaleOnError(t *testing.T) {
	mock := &mockProvider{items: []NewsItem{{ID: 2, Title: "Stale"}}}
	c := NewCache(mock, 50*time.Millisecond)
	c.Start()
	defer c.Stop()

	time.Sleep(20 * time.Millisecond)
	// Переключить на ошибку
	mock.items = nil
	mock.err = errors.New("network error")
	time.Sleep(100 * time.Millisecond)

	// Должен вернуть устаревшие данные
	items := c.Get(10, 0)
	if len(items) != 1 {
		t.Fatalf("expected stale 1 item, got %d", len(items))
	}
}

func TestCacheOffsetAndLimit(t *testing.T) {
	mock := &mockProvider{items: []NewsItem{
		{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5},
	}}
	c := NewCache(mock, time.Hour)
	c.Start()
	defer c.Stop()

	time.Sleep(20 * time.Millisecond)
	items := c.Get(2, 1) // limit=2, offset=1
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
	if items[0].ID != 2 {
		t.Errorf("expected id 2, got %d", items[0].ID)
	}
}
```

**Step 2: Запустить тест — должен упасть**

```bash
go test ./news/... -run TestCache
```

Expected: FAIL

**Step 3: Реализовать**

Создать `news/cache.go`:
```go
package news

import (
	"log"
	"sync"
	"time"
)

const defaultFetchLimit = 100

type Cache struct {
	mu       sync.RWMutex
	provider Provider
	interval time.Duration
	items    []NewsItem
	stop     chan struct{}
}

func NewCache(provider Provider, interval time.Duration) *Cache {
	return &Cache{
		provider: provider,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (c *Cache) Start() {
	c.refresh()
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.refresh()
			case <-c.stop:
				return
			}
		}
	}()
}

func (c *Cache) Stop() {
	close(c.stop)
}

func (c *Cache) refresh() {
	items, err := c.provider.Fetch(defaultFetchLimit)
	if err != nil {
		log.Printf("[news] refresh failed: %v (serving stale cache)", err)
		return
	}
	c.mu.Lock()
	c.items = items
	c.mu.Unlock()
}

func (c *Cache) Get(limit, offset int) []NewsItem {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if offset >= len(c.items) {
		return []NewsItem{}
	}
	end := offset + limit
	if end > len(c.items) {
		end = len(c.items)
	}
	return c.items[offset:end]
}
```

**Step 4: Запустить тест — должен пройти**

```bash
go test ./news/... -run TestCache
```

Expected: PASS

**Step 5: Запустить все тесты news**

```bash
go test ./news/...
```

Expected: все PASS

**Step 6: Commit**

```bash
git add news/cache.go news/cache_test.go
git commit -m "feat: add in-memory news cache with background refresh"
```

---

### Task 6: News handler

**Files:**
- Create: `handlers/news.go`
- Create: `handlers/news_test.go`

**Step 1: Написать тест**

Создать `handlers/news_test.go`:
```go
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
```

**Step 2: Запустить тест — должен упасть**

```bash
go test ./handlers/... -run TestNewsHandler
```

Expected: FAIL

**Step 3: Реализовать**

Создать `handlers/news.go`:
```go
package handlers

import (
	"gml-auth/news"
	"net/http"
	"strconv"
)

const defaultNewsLimit = 10

type NewsCache interface {
	Get(limit, offset int) []news.NewsItem
}

type NewsHandler struct {
	cache NewsCache
}

func NewNewsHandler(cache NewsCache) *NewsHandler {
	return &NewsHandler{cache: cache}
}

func (h *NewsHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	limit := queryInt(r, "limit", defaultNewsLimit)
	offset := queryInt(r, "offset", 0)
	if limit > 100 {
		limit = 100
	}

	items := h.cache.Get(limit, offset)
	if items == nil {
		items = []news.NewsItem{}
	}
	writeJSON(w, http.StatusOK, items)
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return defaultVal
	}
	return n
}
```

**Step 4: Запустить тест — должен пройти**

```bash
go test ./handlers/... -run TestNewsHandler
```

Expected: PASS

**Step 5: Запустить все тесты handlers**

```bash
go test ./handlers/...
```

Expected: все PASS

**Step 6: Commit**

```bash
git add handlers/news.go handlers/news_test.go
git commit -m "feat: add GET /api/news handler"
```

---

### Task 7: Сборка — подключить в main.go

**Files:**
- Modify: `main.go`

**Step 1: Обновить main.go**

Заменить содержимое `main.go`:
```go
package main

import (
	"fmt"
	"gml-auth/config"
	"gml-auth/handlers"
	"gml-auth/news"
	"gml-auth/storage"
	"log"
	"net"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(lw, r)
		log.Printf("[%s] %s %s → %d (%s)",
			r.RemoteAddr, r.Method, r.URL.Path, lw.code, time.Since(start))
	})
}

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

func localIPs() []string {
	var ips []string
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips
}

func buildNewsCache(cfg config.Config) *news.Cache {
	var provider news.Provider
	switch cfg.News.Source {
	case "telegram":
		provider = news.NewTelegramProvider(cfg.News.Token, cfg.News.Channel)
		log.Printf("[news] источник: Telegram канал %s", cfg.News.Channel)
	case "discord":
		provider = news.NewDiscordProvider(cfg.News.Token, cfg.News.Channel)
		log.Printf("[news] источник: Discord канал %s", cfg.News.Channel)
	default:
		log.Printf("[news] источник не настроен (config.json отсутствует или source не задан)")
		return nil
	}
	interval := time.Duration(cfg.News.RefreshSeconds) * time.Second
	cache := news.NewCache(provider, interval)
	cache.Start()
	return cache
}

func main() {
	const port = "5003"

	store := storage.New("data/users.json")
	authHandler := handlers.NewAuthHandler(store)
	adminHandler := handlers.NewAdminHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/", authHandler.SignIn)
	mux.HandleFunc("/api/v1/integrations/auth/signin", authHandler.SignIn)
	mux.HandleFunc("/api/v1/users/refresh", authHandler.Refresh)
	mux.Handle("/admin/users", adminHandler)
	mux.Handle("/admin/users/", adminHandler)

	// Новости — подключаем если есть config.json
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("[news] config.json не найден, /api/news вернёт []")
	}
	newsCache := buildNewsCache(cfg)
	if newsCache != nil {
		newsHandler := handlers.NewNewsHandler(newsCache)
		mux.HandleFunc("/api/news", newsHandler.List)
	} else {
		// Заглушка — возвращает пустой массив
		mux.HandleFunc("/api/news", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]\n"))
		})
	}

	fmt.Println("===========================================")
	fmt.Println("  GML Auth Server")
	fmt.Println("===========================================")
	fmt.Printf("  Порт : %s\n", port)
	for _, ip := range localIPs() {
		fmt.Printf("  Адрес: http://%s:%s\n", ip, port)
	}
	fmt.Printf("  Адрес: http://127.0.0.1:%s  (localhost)\n", port)
	fmt.Println("===========================================")
	log.Printf("Сервер запущен, ожидаю подключения на :%s", port)

	log.Fatal(http.ListenAndServe(":"+port, loggingMiddleware(mux)))
}
```

**Step 2: Собрать**

```bash
go build ./...
```

Expected: без ошибок

**Step 3: Запустить все тесты**

```bash
go test ./...
```

Expected: все PASS

**Step 4: Commit**

```bash
git add main.go
git commit -m "feat: wire news cache into main server"
```

---

### Task 8: Создать пример config.json

**Files:**
- Create: `config.example.json`

**Step 1: Создать файл**

Создать `config.example.json`:
```json
{
  "news": {
    "source": "telegram",
    "token": "1234567890:AABBCCDDEEFFaabbccddeeff1234567890ab",
    "channel": "@your_channel",
    "refresh_seconds": 60
  }
}
```

Для Discord:
```json
{
  "news": {
    "source": "discord",
    "token": "your-discord-bot-token",
    "channel": "1234567890123456789",
    "refresh_seconds": 60
  }
}
```

**Step 2: Добавить config.json в .gitignore**

Создать `.gitignore`:
```
config.json
data/users.json
*.exe
```

**Step 3: Commit**

```bash
git add config.example.json .gitignore
git commit -m "chore: add config.example.json and .gitignore"
```

---

### Task 9: Финальная проверка

**Step 1: Скопировать пример конфига**

```bash
cp config.example.json config.json
# Вписать реальный токен и канал в config.json
```

**Step 2: Запустить сервер**

```bash
go run main.go
```

Expected в логах:
```
[news] источник: Telegram канал @your_channel
GML Auth Server started on :5003
```

**Step 3: Проверить /api/news**

```bash
curl http://localhost:5003/api/news
# Ожидание: JSON массив новостей из канала

curl "http://localhost:5003/api/news?limit=3&offset=0"
# Ожидание: первые 3 новости
```

**Step 4: Финальный commit**

```bash
git add .
git commit -m "feat: complete news integration (Telegram + Discord)"
```
