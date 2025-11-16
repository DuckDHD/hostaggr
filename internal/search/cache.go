package search

import (
	"sync"
	"time"

	"hostaggr/internal/models"
)

type cacheKey struct {
	city    string
	checkin string
	nights  int
	adults  int
}

type cacheEntry struct {
	hotels    []models.Hotel
	expiresAt time.Time
}

type Cache struct {
	mu    sync.RWMutex
	store map[cacheKey]*cacheEntry
	ttl   time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		store: make(map[cacheKey]*cacheEntry),
		ttl:   ttl,
	}

	go c.cleanup()

	return c
}

func (c *Cache) Get(req models.SearchRequest) ([]models.Hotel, bool) {
	key := cacheKey{
		city:    req.City,
		checkin: req.CheckIn,
		nights:  req.Nights,
		adults:  req.Adults,
	}

	c.mu.RLock()
	entry, exists := c.store[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.store, key)
		c.mu.Unlock()
		return nil, false
	}

	return entry.hotels, true
}

func (c *Cache) Set(req models.SearchRequest, hotels []models.Hotel) {
	key := cacheKey{
		city:    req.City,
		checkin: req.CheckIn,
		nights:  req.Nights,
		adults:  req.Adults,
	}

	entry := &cacheEntry{
		hotels:    hotels,
		expiresAt: time.Now().Add(c.ttl),
	}

	c.mu.Lock()
	c.store[key] = entry
	c.mu.Unlock()
}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		c.mu.Lock()
		for key, entry := range c.store {
			if now.After(entry.expiresAt) {
				delete(c.store, key)
			}
		}
		c.mu.Unlock()
	}
}
