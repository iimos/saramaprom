package saramaprom

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// StopFunc represents function for stopping scheduled task.
type StopFunc func()

// Options holds optional params for ExportMetrics.
type Options struct {
	// PrometheusRegistry is prometheus registry. Default prometheus.DefaultRegisterer.
	PrometheusRegistry prometheus.Registerer

	// Namespace and Subsystem form the metric name prefix.
	// Default Subsystem is "sarama".
	Namespace string
	Subsystem string

	// RefreshInterval specifies interval between updating metrics. Default 1s.
	RefreshInterval time.Duration
}

// ExportMetrics exports metrics from go-metrics to prometheus by starting background task,
// which periodically sync sarama metrics to prometheus registry.
func ExportMetrics(metricsRegistry MetricsRegistry, opt Options) StopFunc {
	if opt.PrometheusRegistry == nil {
		opt.PrometheusRegistry = prometheus.DefaultRegisterer
	}
	if opt.Subsystem == "" {
		opt.Subsystem = "sarama"
	}
	if opt.RefreshInterval == 0 {
		opt.RefreshInterval = time.Second
	}

	exp := &exporter{
		opt:                opt,
		registry:           metricsRegistry,
		promRegistry:       opt.PrometheusRegistry,
		gauges:             make(map[string]prometheus.Gauge),
		customMetrics:      make(map[string]*customCollector),
		histogramQuantiles: []float64{0.05, 0.1, 0.25, 0.50, 0.75, 0.9, 0.95, 0.99},
		timerQuantiles:     []float64{0.50, 0.95, 0.99, 0.999},
		mutex:              new(sync.Mutex),
	}

	scheduler := StartScheduler(opt.RefreshInterval, func() {
		err := exp.update()
		if err != nil {
			panic(err)
		}
	})

	return func() {
		_ = exp.unregisterGauges()
		scheduler.Stop()
	}
}

// MetricsRegistry is an interface for 'github.com/rcrowley/go-metrics'.Registry
// which is used for metrics in sarama.
type MetricsRegistry interface {
	Each(func(name string, i interface{}))
}
