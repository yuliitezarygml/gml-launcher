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
	mock.items = nil
	mock.err = errors.New("network error")
	time.Sleep(100 * time.Millisecond)

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
	items := c.Get(2, 1)
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
	if items[0].ID != 2 {
		t.Errorf("expected id 2, got %d", items[0].ID)
	}
}
