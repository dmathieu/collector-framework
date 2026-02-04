package processor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

type Config struct {
	AttributeKey   string `mapstructure:"key"`
	AttributeValue string `mapstructure:"value"`
}

func createTracesProcessor(ctx context.Context, set processor.Settings, cfg component.Config, nextConsumer consumer.Traces) (processor.Traces, error) {
	conf := cfg.(*Config)

	return processorhelper.NewTraces(ctx, set, conf, nextConsumer, func(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
		rs := td.ResourceSpans()

		for i := 0; i < rs.Len(); i++ {
			ils := rs.At(i).ScopeSpans()
			for j := 0; j < ils.Len(); j++ {
				spans := ils.At(j).Spans()
				for k := 0; k < spans.Len(); k++ {
					span := spans.At(k)
					span.Attributes().PutStr(conf.AttributeKey, conf.AttributeValue)
				}
			}
		}

		return td, nil
	})
}

func NewFactory() processor.Factory {
	return processor.NewFactory(
		component.MustNewType("addattribute"),
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		AttributeKey:   "env",
		AttributeValue: "default",
	}
}
