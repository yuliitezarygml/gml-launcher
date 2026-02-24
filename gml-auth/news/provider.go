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
