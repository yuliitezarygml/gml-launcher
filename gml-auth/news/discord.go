package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
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
	for i, msg := range messages {
		if msg.Content == "" {
			continue
		}
		title, _ := splitTitle(msg.Content)

		var createdAt string
		if t, err := time.Parse(time.RFC3339Nano, strings.TrimSuffix(msg.Timestamp, "+00:00")+"Z"); err == nil {
			createdAt = t.UTC().Format(time.RFC3339)
		} else {
			createdAt = msg.Timestamp
		}

		items = append(items, NewsItem{
			ID:          i + 1,
			Title:       title,
			Description: msg.Content,
			CreatedAt:   createdAt,
		})
	}
	return items, nil
}
