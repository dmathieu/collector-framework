package connector

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type Config struct {
	MetricName string `mapstructure:"metric_name"`
}

type spanMetricsConnector struct {
	mu              sync.Mutex
	config          *Config
	metricsConsumer consumer.Metrics
	counts          map[string]int64
}

func (c *spanMetricsConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (c *spanMetricsConnector) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < td.ResourceSpans().Len(); i++ {
		rs := td.ResourceSpans().At(i)
		serviceName := "unknown"
		if sn, ok := rs.Resource().Attributes().Get("service.name"); ok {
			serviceName = sn.Str()
		}

		spanCount := 0
		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			spanCount += rs.ScopeSpans().At(j).Spans().Len()
		}

		c.counts[serviceName] += int64(spanCount)
	}

	metrics := c.generateMetrics()
	return c.metricsConsumer.ConsumeMetrics(ctx, metrics)
}

func (c *spanMetricsConnector) generateMetrics() pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()

	metric := sm.Metrics().AppendEmpty()
	metric.SetName(c.config.MetricName)
	metric.SetDescription("Count of spans by service")

	sum := metric.SetEmptySum()
	sum.SetIsMonotonic(true)
	sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

	for service, count := range c.counts {
		dp := sum.DataPoints().AppendEmpty()
		dp.SetIntValue(count)
		dp.Attributes().PutStr("service.name", service)
		dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	}

	return metrics
}

func (c *spanMetricsConnector) Start(ctx context.Context, host component.Host) error {
	return nil
}

func (c *spanMetricsConnector) Shutdown(ctx context.Context) error {
	return nil
}

func NewFactory() connector.Factory {
	return connector.NewFactory(
		component.MustNewType("spanmetrics"),
		createDefaultConfig,
		connector.WithTracesToMetrics(createTracesToMetrics, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		MetricName: "span_count",
	}
}

func createTracesToMetrics(ctx context.Context, set connector.Settings, cfg component.Config, metrics consumer.Metrics) (connector.Traces, error) {
	conf := cfg.(*Config)

	return &spanMetricsConnector{
		config:          conf,
		metricsConsumer: metrics,
		counts:          make(map[string]int64),
	}, nil
}
