package config

import (
	"encoding/json"
	"os"
)

type TelegramConfig struct {
	Token   string `json:"token"`
	Channel string `json:"channel"`
}

type DiscordConfig struct {
	Token   string `json:"token"`
	Channel string `json:"channel"`
}

type NewsConfig struct {
	RefreshSeconds int            `json:"refresh_seconds"`
	Telegram       TelegramConfig `json:"telegram"`
	Discord        DiscordConfig  `json:"discord"`
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
