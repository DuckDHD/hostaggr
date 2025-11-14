package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hostaggr/internal/models"
	"hostaggr/internal/search"
)

type Handler struct {
	aggregator  *search.Aggregator
	rateLimiter *search.RateLimiter
}

func NewHandler(agg *search.Aggregator, rl *search.RateLimiter) *Handler {
	return &Handler{
		aggregator:  agg,
		rateLimiter: rl,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

type healthResponse struct {
	Status string `json:"status"`
}

// SearchHotels handles GET /search requests
func (h *Handler) SearchHotels(w http.ResponseWriter, r *http.Request) {
	// Extract IP address
	ip := extractIP(r)

	// Check rate limit
	if !h.rateLimiter.Allow(ip) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "rate limit exceeded",
		})
		return
	}

	// Parse and validate query parameters
	city := r.URL.Query().Get("city")
	if city == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "city parameter is required",
		})
		return
	}

	checkin := r.URL.Query().Get("checkin")
	if checkin == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "checkin parameter is required",
		})
		return
	}

	// Validate checkin format (YYYY-MM-DD)
	if !isValidDateFormat(checkin) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "checkin must be in YYYY-MM-DD format",
		})
		return
	}

	nightsStr := r.URL.Query().Get("nights")
	if nightsStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "nights parameter is required",
		})
		return
	}

	nights, err := strconv.Atoi(nightsStr)
	if err != nil || nights <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "nights must be a positive integer",
		})
		return
	}

	adultsStr := r.URL.Query().Get("adults")
	if adultsStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "adults parameter is required",
		})
		return
	}

	adults, err := strconv.Atoi(adultsStr)
	if err != nil || adults <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "adults must be a positive integer",
		})
		return
	}

	// Create search request
	req := models.SearchRequest{
		City:    city,
		CheckIn: checkin,
		Nights:  nights,
		Adults:  adults,
	}

	// Create context with 5-second timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Perform search
	response, err := h.aggregator.Search(ctx, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "internal server error",
		})
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Health handles GET /healthz requests
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthResponse{
		Status: "ok",
	})
}

// Metrics handles GET /metrics requests
func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	// Placeholder for metrics - will be implemented when obs.Metrics is created
	metrics := map[string]interface{}{
		"requests_total":  0,
		"cache_hits":      0,
		"cache_misses":    0,
		"provider_errors": 0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

func extractIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	ip := r.RemoteAddr
	// RemoteAddr includes port, strip it
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

func isValidDateFormat(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
