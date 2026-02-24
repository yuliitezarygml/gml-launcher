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
