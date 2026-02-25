package config

import (
	"os"
	"testing"
)

func TestLoadConfigBoth(t *testing.T) {
	f, err := os.CreateTemp("", "cfg-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{
		"news": {
			"refresh_seconds": 30,
			"telegram": {"token": "tg-tok", "channel": "@test"},
			"discord":  {"token": "dc-tok", "channel": "123456"}
		}
	}`); err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.News.Telegram.Token != "tg-tok" {
		t.Errorf("telegram token: got %s", cfg.News.Telegram.Token)
	}
	if cfg.News.Discord.Token != "dc-tok" {
		t.Errorf("discord token: got %s", cfg.News.Discord.Token)
	}
	if cfg.News.RefreshSeconds != 30 {
		t.Errorf("refresh_seconds: got %d", cfg.News.RefreshSeconds)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	f, err := os.CreateTemp("", "cfg-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"news":{"telegram":{"token":"tok","channel":"@ch"}}}`); err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatal(err)
	}
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
