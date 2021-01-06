# saramaprom
[![GoDoc](https://godoc.org/github.com/iimos/saramaprom?status.png)](http://godoc.org/github.com/iimos/saramaprom)
[![Go Report](https://goreportcard.com/badge/github.com/iimos/saramaprom)](https://goreportcard.com/report/github.com/iimos/saramaprom)

This is a prometheus metrics reporter for the [sarama](https://github.com/Shopify/sarama) library. 
It is based on https://github.com/deathowl/go-metrics-prometheus library.

## Why
Because `go-metrics-prometheus` is a general solution it reports metrics with no labels so it's hard to use. Thus a sarama specific solution was made, it reports metrics with labels for brokers, topics and consumer/producer instance.

## Installation
```console
go get github.com/iimos/saramaprom
```

## Usage

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

Metric names by default:
```
Gauges:
sarama_batch_size
sarama_compression_ratio
sarama_incoming_byte_rate
sarama_outgoing_byte_rate
sarama_record_send_rate
sarama_records_per_request
sarama_request_latency_in_ms
sarama_request_rate
sarama_request_size
sarama_requests_in_flight
sarama_response_rate
sarama_response_size

Histograms:
sarama_batch_size_histogram
sarama_compression_ratio_histogram
sarama_records_per_request_histogram
sarama_request_latency_in_ms_histogram
sarama_request_size_histogram
sarama_response_size_histogram
```

Every metric have three labels:
* broker – kafka broker id
* topic – kafka topic name
* label – custom label to distinguish different consumers/producers


## Requirements

Go 1.13 or above.
