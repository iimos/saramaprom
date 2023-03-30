package saramaprom

// This code is based on a code of https://github.com/deathowl/go-metrics-prometheus library.

import (
	"fmt"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
)

type exporter struct {
	opt                Options
	registry           MetricsRegistry
	promRegistry       prometheus.Registerer
	gauges             map[string]prometheus.Gauge
	customMetrics      map[string]*customCollector
	histogramQuantiles []float64
	timerQuantiles     []float64
	mutex              *sync.Mutex
}

func (c *exporter) sanitizeName(key string) string {
	ret := []byte(key)
	for i := 0; i < len(ret); i++ {
		c := key[i]
		allowed := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == ':' || (c >= '0' && c <= '9')
		if !allowed {
			ret[i] = '_'
		}
	}
	return string(ret)
}

func (c *exporter) createKey(name string) string {
	return c.opt.Namespace + "_" + c.opt.Subsystem + "_" + name
}

func (c *exporter) gaugeFromNameAndValue(name string, val float64) error {
	shortName, labels, skip := c.metricNameAndLabels(name)
	if skip {
		return nil
	}

	if _, exists := c.gauges[name]; !exists {
		labelNames := make([]string, 0, len(labels))
		for labelName := range labels {
			labelNames = append(labelNames, labelName)
		}

		g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: c.sanitizeName(c.opt.Namespace),
			Subsystem: c.sanitizeName(c.opt.Subsystem),
			Name:      c.sanitizeName(shortName),
		}, labelNames)

		if err := c.promRegistry.Register(g); err != nil {
			switch err := err.(type) {
			case prometheus.AlreadyRegisteredError:
				var ok bool
				g, ok = err.ExistingCollector.(*prometheus.GaugeVec)
				if !ok {
					return fmt.Errorf("prometheus collector already registered but it's not *prometheus.GaugeVec: %v", g)
				}
			default:
				return err
			}
		}
		c.gauges[name] = g.With(labels)
	}

	c.gauges[name].Set(val)
	return nil
}

// unregisterGauges will remove the gauge metrics so that they do not show
// incorrect values after the application has been shut down.
func (c *exporter) unregisterGauges() error {
	for _, g := range c.gauges {
		if ok := c.promRegistry.Unregister(g); !ok {
			return fmt.Errorf("unable to unregister prometheus collector")
		}
	}

	return nil
}

func (c *exporter) metricNameAndLabels(metricName string) (newName string, labels map[string]string, skip bool) {
	newName, broker, topic := parseMetricName(metricName)
	if broker == "" && topic == "" {
		// skip metrics for total
		return newName, labels, true
	}
	labels = map[string]string{
		"broker": broker,
		"topic":  topic,
	}
	return newName, labels, false
}

func parseMetricName(name string) (newName, broker, topic string) {
	if i := strings.Index(name, "-for-broker-"); i >= 0 {
		newName = name[:i]
		broker = name[i+len("-for-broker-"):]
		return
	}
	if i := strings.Index(name, "-for-topic-"); i >= 0 {
		newName = name[:i]
		topic = name[i+len("-for-topic-"):]
		return
	}
	return name, "", ""
}

func (c *exporter) summaryFromNameAndMetric(name string, goMetric interface{}, quantiles []float64) error {
	key := c.createKey(name)
	collector, exists := c.customMetrics[key]
	if !exists {
		collector = newCustomCollector(c.mutex)
		c.promRegistry.MustRegister(collector)
		c.customMetrics[key] = collector
	}

	var ps []float64
	var count uint64
	var sum float64

	switch metric := goMetric.(type) {
	case metrics.Histogram:
		snapshot := metric.Snapshot()
		ps = snapshot.Percentiles(quantiles)
		count = uint64(snapshot.Count())
		sum = float64(snapshot.Sum())
	case metrics.Timer:
		snapshot := metric.Snapshot()
		ps = snapshot.Percentiles(quantiles)
		count = uint64(snapshot.Count())
		sum = float64(snapshot.Sum())
	default:
		return fmt.Errorf("unexpected metric type %T", goMetric)
	}

	quantilesVals := make(map[float64]float64)
	for i, quantile := range quantiles {
		quantilesVals[quantile] = ps[i]
	}

	name, labels, skip := c.metricNameAndLabels(name)
	if skip {
		return nil
	}

	desc := prometheus.NewDesc(
		prometheus.BuildFQName(
			c.sanitizeName(c.opt.Namespace),
			c.sanitizeName(c.opt.Subsystem),
			c.sanitizeName(name),
		),
		c.sanitizeName(name),
		nil,
		labels,
	)

	collector.setMetric(prometheus.MustNewConstSummary(desc, count, sum, quantilesVals))

	return nil
}

func (c *exporter) update() error {
	var err error
	c.registry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			err = c.gaugeFromNameAndValue(name, float64(metric.Count()))
		case metrics.Gauge:
			err = c.gaugeFromNameAndValue(name, float64(metric.Value()))
		case metrics.GaugeFloat64:
			err = c.gaugeFromNameAndValue(name, metric.Value())
		case metrics.Histogram:
			err = c.summaryFromNameAndMetric(name, metric, c.histogramQuantiles)
		case metrics.Meter:
			lastSample := metric.Snapshot().Rate1()
			err = c.gaugeFromNameAndValue(name, lastSample)
		case metrics.Timer:
			err = c.summaryFromNameAndMetric(name, metric, c.timerQuantiles)
		}
	})
	return err
}

// customCollector is a collector of prometheus.constSummary objects.
type customCollector struct {
	prometheus.Collector

	metric prometheus.Metric
	mutex  *sync.Mutex
}

func newCustomCollector(mu *sync.Mutex) *customCollector {
	return &customCollector{
		mutex: mu,
	}
}

func (c *customCollector) setMetric(metric prometheus.Metric) {
	c.mutex.Lock()
	c.metric = metric
	c.mutex.Unlock()
}

func (c *customCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	if c.metric != nil {
		val := c.metric
		ch <- val
	}
	c.mutex.Unlock()
}

func (c *customCollector) Describe(_ chan<- *prometheus.Desc) {
	// empty method to fulfill prometheus.Collector interface
}
