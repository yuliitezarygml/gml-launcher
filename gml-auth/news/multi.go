package news

import (
	"log"
	"sort"
)

// MultiProvider объединяет несколько источников новостей в одну ленту
type MultiProvider struct {
	providers []Provider
}

func NewMultiProvider(providers ...Provider) *MultiProvider {
	return &MultiProvider{providers: providers}
}

func (m *MultiProvider) Fetch(limit int) ([]NewsItem, error) {
	var all []NewsItem
	for _, p := range m.providers {
		items, err := p.Fetch(limit)
		if err != nil {
			log.Printf("[news] provider error: %v", err)
			continue
		}
		all = append(all, items...)
	}
	// Сортировка по дате убывания (RFC3339 сортируется лексикографически)
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt > all[j].CreatedAt
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}
