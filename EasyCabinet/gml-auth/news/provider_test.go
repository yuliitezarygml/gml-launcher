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
