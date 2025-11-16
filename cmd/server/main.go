package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"hostaggr/internal/config"
	httphandler "hostaggr/internal/http"
	custommiddleware "hostaggr/internal/middleware"
	"hostaggr/internal/obs"
	"hostaggr/internal/providers"
	"hostaggr/internal/search"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting hotel aggregator server...")

	logger.Info("Loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("Initializing providers...")
	providersList := []providers.Provider{
		providers.NewMock1(),
		providers.NewMock2(),
		providers.NewMock3(),
	}
	logger.Info("Initialized providers", "count", len(providersList))

	logger.Info("Initializing cache...")
	cache := search.NewCache(cfg.CacheTTL)

	logger.Info("Initializing metrics...")
	metrics := obs.NewMetrics()

	logger.Info("Initializing aggregator...")
	aggregator := search.NewAggregator(providersList, cache, cfg.ProviderTimeout, metrics, logger)

	logger.Info("Initializing rate limiter...")
	rateLimiter := search.NewRateLimiter(cfg.RateLimitMaxTokens, cfg.RateLimitRefillInterval)

	logger.Info("Initializing HTTP handler...")
	handler := httphandler.NewHandler(aggregator, rateLimiter, metrics, logger)

	logger.Info("Setting up router...")
	router := setupRouter(handler, logger)

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server error", "error", err)
		os.Exit(1)

	case sig := <-shutdown:
		logger.Info("Received signal, starting graceful shutdown", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Error during shutdown", "error", err)
			if err := server.Close(); err != nil {
				logger.Error("Error forcing server close", "error", err)
			}
		}

		logger.Info("Server stopped gracefully")
	}
}

func setupRouter(h *httphandler.Handler, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(custommiddleware.RequestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/search", h.SearchHotels)
	r.Get("/healthz", h.Health)
	r.Get("/metrics", h.Metrics)

	return r
}
