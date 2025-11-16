package obs

import (
	"fmt"
	"sort"
	"strings"
)

func (m *Metrics) ToPrometheusFormat() string {
	snapshot := m.GetSnapshot()
	
	var sb strings.Builder
	
	sb.WriteString("# HELP hostaggr_requests_total Total number of search requests received\n")
	sb.WriteString("# TYPE hostaggr_requests_total counter\n")
	sb.WriteString(fmt.Sprintf("hostaggr_requests_total %d\n", snapshot.RequestsTotal))
	sb.WriteString("\n")
	
	sb.WriteString("# HELP hostaggr_cache_hits_total Number of requests served from cache\n")
	sb.WriteString("# TYPE hostaggr_cache_hits_total counter\n")
	sb.WriteString(fmt.Sprintf("hostaggr_cache_hits_total %d\n", snapshot.CacheHits))
	sb.WriteString("\n")
	
	sb.WriteString("# HELP hostaggr_cache_misses_total Number of requests that required provider queries\n")
	sb.WriteString("# TYPE hostaggr_cache_misses_total counter\n")
	sb.WriteString(fmt.Sprintf("hostaggr_cache_misses_total %d\n", snapshot.CacheMisses))
	sb.WriteString("\n")
	
	sb.WriteString("# HELP hostaggr_provider_requests_total Total number of requests per provider by status\n")
	sb.WriteString("# TYPE hostaggr_provider_requests_total counter\n")
	
	providerNames := make(map[string]bool)
	for provider := range snapshot.ProviderSuccesses {
		providerNames[provider] = true
	}
	for provider := range snapshot.ProviderErrors {
		providerNames[provider] = true
	}
	
	sortedProviders := make([]string, 0, len(providerNames))
	for provider := range providerNames {
		sortedProviders = append(sortedProviders, provider)
	}
	sort.Strings(sortedProviders)
	
	for _, provider := range sortedProviders {
		successCount := snapshot.ProviderSuccesses[provider]
		sb.WriteString(fmt.Sprintf("hostaggr_provider_requests_total{provider=\"%s\",status=\"success\"} %d\n", provider, successCount))
	}
	
	for _, provider := range sortedProviders {
		errorCount := snapshot.ProviderErrors[provider]
		sb.WriteString(fmt.Sprintf("hostaggr_provider_requests_total{provider=\"%s\",status=\"error\"} %d\n", provider, errorCount))
	}
	
	return sb.String()
}
