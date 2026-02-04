package receiver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
)

type Config struct {
	Port int `mapstructure:"port"`
}

type httpReceiver struct {
	config   *Config
	consumer consumer.Traces
	server   *http.Server
}

func (r *httpReceiver) Start(ctx context.Context, host component.Host) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/traces", r.handleTraces)

	r.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", r.config.Port),
		Handler: mux,
	}

	go func() {
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	return nil
}

func (r *httpReceiver) Shutdown(ctx context.Context) error {
	if r.server != nil {
		return r.server.Shutdown(ctx)
	}
	return nil
}

func (r *httpReceiver) handleTraces(w http.ResponseWriter, req *http.Request) {
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	ss := rs.ScopeSpans().AppendEmpty()
	span := ss.Spans().AppendEmpty()

	span.SetName("example.span")
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(time.Now().Add(time.Second)))

	if err := r.consumer.ConsumeTraces(req.Context(), traces); err != nil {
		http.Error(w, "Failed to process traces", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("customhttp"),
		createDefaultConfig,
		receiver.WithTraces(createTracesReceiver, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Port: 2504,
	}
}

func createTracesReceiver(ctx context.Context, set receiver.Settings, cfg component.Config, consumer consumer.Traces) (receiver.Traces, error) {
	conf := cfg.(*Config)

	return &httpReceiver{
		config:   conf,
		consumer: consumer,
	}, nil
}
