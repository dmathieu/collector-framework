package scraper

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`
}

func createDefaultConfig() component.Config {
	return &Config{
		ControllerConfig: scraperhelper.NewDefaultControllerConfig(),
	}
}

type cpuScraper struct {
	config *Config
}

func (s *cpuScraper) start(ctx context.Context, host component.Host) error {
	return nil
}

func (s *cpuScraper) shutdown(ctx context.Context) error {
	return nil
}

func (s *cpuScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()

	cpuPercent := 45.5 // Turn this into actual CPU usage

	metric := sm.Metrics().AppendEmpty()
	metric.SetName("system.cpu.utilization")
	metric.SetDescription("CPU utilization")
	metric.SetUnit("1")

	gauge := metric.SetEmptyGauge()
	dp := gauge.DataPoints().AppendEmpty()
	dp.SetDoubleValue(cpuPercent)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	return metrics, nil
}

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("cpuscraper"),
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelDevelopment),
	)
}

func createMetricsReceiver(ctx context.Context, set receiver.Settings, cfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	conf := cfg.(*Config)

	s := &cpuScraper{config: conf}
	ms, err := scraper.NewMetrics(
		s.scrape,
		scraper.WithStart(s.start),
		scraper.WithShutdown(s.shutdown),
	)
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewMetricsController(
		&conf.ControllerConfig,
		set,
		consumer,
		scraperhelper.AddMetricsScraper(component.MustNewType("cpuscraper"), ms),
	)
}
