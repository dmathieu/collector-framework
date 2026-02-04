package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type Config struct {
	Endpoint string `mapstructure:"endpoint"`
}

type metricsExporter struct {
	config *Config
}

func (e *metricsExporter) Start(ctx context.Context, host component.Host) error {
	return nil
}

func (e *metricsExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (e *metricsExporter) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	payload := convertMetricsToJSON(md)

	req, err := http.NewRequestWithContext(ctx, "POST", e.config.Endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %d", resp.StatusCode)
	}

	return nil
}

func convertMetricsToJSON(md pmetric.Metrics) []byte {
	data := make([]map[string]interface{}, 0)

	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				data = append(data, map[string]interface{}{
					"name": metric.Name(),
					"type": metric.Type().String(),
				})
			}
		}
	}

	result, _ := json.Marshal(data)
	return result
}

func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		component.MustNewType("customapi"),
		createDefaultConfig,
		exporter.WithMetrics(createMetricsExporter, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Endpoint: "http://localhost:1304/metrics",
	}
}

func createMetricsExporter(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Metrics, error) {
	conf := cfg.(*Config)
	exp := &metricsExporter{config: conf}
	return exporterhelper.NewMetrics(ctx, set, cfg, exp.pushMetrics)
}
