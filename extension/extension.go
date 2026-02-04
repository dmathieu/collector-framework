package extension

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

type Config struct {
	Port int `mapstructure:"port"`
}

type healthCheckExtension struct {
	config *Config
	server *http.Server
	ready  atomic.Bool
}

func (h *healthCheckExtension) Start(ctx context.Context, host component.Host) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("/readyz", h.handleReadiness)

	h.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", h.config.Port),
		Handler: mux,
	}

	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Health check server failed: %v", err)
		}
	}()
	h.ready.Store(true)

	return nil
}

func (h *healthCheckExtension) Shutdown(ctx context.Context) error {
	h.ready.Store(false)
	if h.server != nil {
		return h.server.Shutdown(ctx)
	}
	return nil
}

func (h *healthCheckExtension) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (h *healthCheckExtension) handleReadiness(w http.ResponseWriter, r *http.Request) {
	if h.ready.Load() {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Ready"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("Not Ready"))
}

func NewFactory() extension.Factory {
	return extension.NewFactory(
		component.MustNewType("healthcheck"),
		createDefaultConfig,
		createExtension,
		component.StabilityLevelDevelopment,
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Port: 2806,
	}
}

func createExtension(ctx context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
	conf := cfg.(*Config)
	return &healthCheckExtension{config: conf}, nil
}
