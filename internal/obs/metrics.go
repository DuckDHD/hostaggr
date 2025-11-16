package obs

import (
	"sync"
)

type Metrics struct {
	mu                sync.RWMutex
	requestsTotal     int64
	cacheHits         int64
	cacheMisses       int64
	providerErrors    map[string]int64
	providerSuccesses map[string]int64
}

type MetricsSnapshot struct {
	RequestsTotal     int64            `json:"requests_total"`
	CacheHits         int64            `json:"cache_hits"`
	CacheMisses       int64            `json:"cache_misses"`
	ProviderErrors    map[string]int64 `json:"provider_errors"`
	ProviderSuccesses map[string]int64 `json:"provider_successes"`
}

func NewMetrics() *Metrics {
	return &Metrics{
		providerErrors:    make(map[string]int64),
		providerSuccesses: make(map[string]int64),
	}
}

func (m *Metrics) IncrementRequests() {
	m.mu.Lock()
	m.requestsTotal++
	m.mu.Unlock()
}

func (m *Metrics) IncrementCacheHits() {
	m.mu.Lock()
	m.cacheHits++
	m.mu.Unlock()
}

func (m *Metrics) IncrementCacheMisses() {
	m.mu.Lock()
	m.cacheMisses++
	m.mu.Unlock()
}

func (m *Metrics) IncrementProviderError(providerName string) {
	m.mu.Lock()
	m.providerErrors[providerName]++
	m.mu.Unlock()
}

func (m *Metrics) IncrementProviderSuccess(providerName string) {
	m.mu.Lock()
	m.providerSuccesses[providerName]++
	m.mu.Unlock()
}

func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	errors := make(map[string]int64, len(m.providerErrors))
	for k, v := range m.providerErrors {
		errors[k] = v
	}

	successes := make(map[string]int64, len(m.providerSuccesses))
	for k, v := range m.providerSuccesses {
		successes[k] = v
	}

	return MetricsSnapshot{
		RequestsTotal:     m.requestsTotal,
		CacheHits:         m.cacheHits,
		CacheMisses:       m.cacheMisses,
		ProviderErrors:    errors,
		ProviderSuccesses: successes,
	}
}
