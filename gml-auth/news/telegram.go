package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const telegramBaseURL = "https://api.telegram.org"

func splitTitle(text string) (title, rest string) {
	parts := strings.SplitN(text, "\n", 2)
	title = parts[0]
	if len([]rune(title)) > 100 {
		title = string([]rune(title)[:100])
	}
	if len(parts) > 1 {
		rest = parts[1]
	}
	return
}

type tgChat struct {
	Username string `json:"username"`
}

type tgMessage struct {
	MessageID int    `json:"message_id"`
	Chat      tgChat `json:"chat"`
	Text      string `json:"text"`
	Date      int64  `json:"date"`
}

type tgUpdate struct {
	UpdateID    int        `json:"update_id"`
	ChannelPost *tgMessage `json:"channel_post"`
}

type telegramResponse struct {
	OK     bool       `json:"ok"`
	Result []tgUpdate `json:"result"`
}

type TelegramProvider struct {
	token   string
	channel string // e.g. "@sinkdev_dev"
	baseURL string
	mu      sync.Mutex
	offset  int
	stored  []NewsItem
}

func NewTelegramProvider(token, channel string) *TelegramProvider {
	return &TelegramProvider{
		token:   token,
		channel: strings.TrimPrefix(channel, "@"),
		baseURL: telegramBaseURL,
	}
}

func (p *TelegramProvider) Fetch(limit int) ([]NewsItem, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	url := fmt.Sprintf("%s/bot%s/getUpdates?limit=100&allowed_updates=[\"channel_post\"]&offset=%d",
		p.baseURL, p.token, p.offset)

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return p.top(limit), err
	}
	defer resp.Body.Close()

	var tgResp telegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return p.top(limit), err
	}

	for _, upd := range tgResp.Result {
		if upd.UpdateID >= p.offset {
			p.offset = upd.UpdateID + 1
		}
		msg := upd.ChannelPost
		if msg == nil || msg.Text == "" {
			continue
		}
		chanName := strings.TrimPrefix(p.channel, "@")
		if !strings.EqualFold(msg.Chat.Username, chanName) {
			continue
		}
		title, _ := splitTitle(msg.Text)
		p.stored = append(p.stored, NewsItem{
			ID:          msg.MessageID,
			Title:       title,
			Description: msg.Text,
			CreatedAt:   time.Unix(msg.Date, 0).UTC().Format(time.RFC3339),
		})
	}

	return p.top(limit), nil
}

// top returns the last `limit` stored items (most recent first).
func (p *TelegramProvider) top(limit int) []NewsItem {
	all := p.stored
	if len(all) > limit {
		all = all[len(all)-limit:]
	}
	// return in reverse (newest last → newest first)
	out := make([]NewsItem, len(all))
	for i, item := range all {
		out[len(all)-1-i] = item
	}
	return out
}
