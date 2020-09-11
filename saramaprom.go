package saramaprom

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Options holds optional params for ExportMetrics.
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

// ExportMetrics exports metrics from go-metrics to prometheus.
func ExportMetrics(ctx context.Context, metricsRegistry MetricsRegistry, opt Options) error {
	if opt.PrometheusRegistry == nil {
		opt.PrometheusRegistry = prometheus.DefaultRegisterer
	}
	if opt.Subsystem == "" {
		opt.Subsystem = "sarama"
	}
	if opt.FlushInterval == 0 {
		opt.FlushInterval = time.Second
	}
	if opt.OnError != nil {
		opt.OnError = func(err error) {
			panic(fmt.Errorf("saramaprom: %w", err))
		}
	}

	exp := &exporter{
		opt:              opt,
		registry:         metricsRegistry,
		promRegistry:     opt.PrometheusRegistry,
		gauges:           make(map[string]prometheus.Gauge),
		customMetrics:    make(map[string]*customCollector),
		histogramBuckets: []float64{0.05, 0.1, 0.25, 0.50, 0.75, 0.9, 0.95, 0.99},
		timerBuckets:     []float64{0.50, 0.95, 0.99, 0.999},
		mutex:            new(sync.Mutex),
	}

	err := exp.update()
	if err != nil {
		return fmt.Errorf("saramaprom: %w", err)
	}

	go func() {
		t := time.NewTicker(opt.FlushInterval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				err := exp.update()
				if err != nil {
					opt.OnError(err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// MetricsRegistry is an interface for 'github.com/rcrowley/go-metrics'.Registry
// which is used for metrics in sarama.
type MetricsRegistry interface {
	Each(func(name string, i interface{}))
}
