package search

import (
	"sync"
	"time"

	"hostaggr/internal/models"
)

// cacheKey represents the unique identifier for a search request
type cacheKey struct {
	city    string
	checkin string
	nights  int
	adults  int
}

// cacheEntry stores cached hotels with an expiration timestamp
type cacheEntry struct {
	hotels    []models.Hotel
	expiresAt time.Time
}

// Cache provides thread-safe in-memory caching for hotel search results
type Cache struct {
	mu    sync.RWMutex
	store map[cacheKey]*cacheEntry
	ttl   time.Duration
}

// NewCache creates a new cache with the specified TTL and starts a background cleanup goroutine
func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		store: make(map[cacheKey]*cacheEntry),
		ttl:   ttl,
	}

	// Start background cleanup goroutine
	go c.cleanup()

	return c
}

// Get retrieves cached hotels for a search request
// Returns the hotels and true if found and not expired, otherwise nil and false
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

	// Check if entry has expired
	if time.Now().After(entry.expiresAt) {
		// Entry expired, remove it
		c.mu.Lock()
		delete(c.store, key)
		c.mu.Unlock()
		return nil, false
	}

	return entry.hotels, true
}

// Set stores hotels in the cache for a search request with a 30-second TTL
func (c *Cache) Set(req models.SearchRequest, hotels []models.Hotel) {
	key := cacheKey{
		city:    req.City,
		checkin: req.CheckIn,
		nights:  req.Nights,
		adults:  req.Adults,
	}

	entry := &cacheEntry{
		hotels:    hotels,
		expiresAt: time.Now().Add(30 * time.Second),
	}

	c.mu.Lock()
	c.store[key] = entry
	c.mu.Unlock()
}

// cleanup runs in the background and removes expired entries every 60 seconds
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
