package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hostaggr/internal/middleware"
	"hostaggr/internal/models"
	"hostaggr/internal/obs"
	"hostaggr/internal/search"
)

type Handler struct {
	aggregator  *search.Aggregator
	rateLimiter *search.RateLimiter
	metrics     *obs.Metrics
	logger      *slog.Logger
}

func NewHandler(agg *search.Aggregator, rl *search.RateLimiter, m *obs.Metrics, logger *slog.Logger) *Handler {
	return &Handler{
		aggregator:  agg,
		rateLimiter: rl,
		metrics:     m,
		logger:      logger,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

type healthResponse struct {
	Status string `json:"status"`
}

func (h *Handler) SearchHotels(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	ip := extractIP(r)

	if !h.rateLimiter.Allow(ip) {
		if h.logger != nil {
			h.logger.Warn("rate limit exceeded",
				"request_id", requestID,
				"ip", ip,
			)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "rate limit exceeded",
		})
		return
	}

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

	req := models.SearchRequest{
		City:    city,
		CheckIn: checkin,
		Nights:  nights,
		Adults:  adults,
	}

	h.metrics.IncrementRequests()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response, err := h.aggregator.Search(ctx, req)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("search failed",
				"request_id", requestID,
				"error", err.Error(),
			)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "internal server error",
		})
		return
	}

	if response.Stats.Cache == "hit" {
		h.metrics.IncrementCacheHits()
	} else {
		h.metrics.IncrementCacheMisses()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthResponse{
		Status: "ok",
	})
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	acceptHeader := r.Header.Get("Accept")

	if format == "prometheus" || strings.Contains(acceptHeader, "text/plain") {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(h.metrics.ToPrometheusFormat()))
		return
	}

	snapshot := h.metrics.GetSnapshot()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(snapshot)
}

func extractIP(r *http.Request) string {
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func isValidDateFormat(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
