package news

import (
	"sync"
	"time"
)

// Cache wraps a Provider and periodically refreshes its data,
// serving stale data when the provider returns an error.
type Cache struct {
	provider Provider
	interval time.Duration

	mu    sync.RWMutex
	items []NewsItem

	stop chan struct{}
	done chan struct{}
}

func NewCache(p Provider, interval time.Duration) *Cache {
	return &Cache{
		provider: p,
		interval: interval,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (c *Cache) Start() {
	go func() {
		defer close(c.done)
		// Initial fetch
		c.refresh()
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
	<-c.done
}

func (c *Cache) refresh() {
	items, err := c.provider.Fetch(100)
	if err != nil {
		// keep stale data
		return
	}
	c.mu.Lock()
	c.items = items
	c.mu.Unlock()
}

// Get returns up to limit items starting at offset.
func (c *Cache) Get(limit, offset int) []NewsItem {
	c.mu.RLock()
	all := c.items
	c.mu.RUnlock()

	if offset >= len(all) {
		return []NewsItem{}
	}
	all = all[offset:]
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all
}
