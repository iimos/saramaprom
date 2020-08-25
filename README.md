# saramaprom
This is a prometheus metrics reporter for the sarama library. 
It is based on is based on https://github.com/deathowl/go-metrics-prometheus library.

Usage:

```go
import (
    "context"
    "github.com/Shopify/sarama"
    "github.com/iimos/saramaprom"
)

ctx := context.Background()
cfg := sarama.NewConfig()
err := saramaprom.ExportMetrics(ctx, cfg.MetricRegistry, saramaprom.Options{})
```

Posible options:
```go
type Options struct {
	// PrometheusRegistry is prometheus registry. Default prometheus.DefaultRegisterer.
	PrometheusRegistry prometheus.Registerer

	// Namespace and Subsystem form the metric name prefix.
	// Default Subsystem is "sarama".
	Namespace string
	Subsystem string

	// Label specifies value of "label" label. Default "".
	Label string

	// FlushInterval specifies interval between updating metrics. Default 1s.
	FlushInterval time.Duration

	// OnError is error handler. Default handler panics when error occurred.
	OnError func(err error)

	// Debug turns on debug logging.
	Debug bool
}
```