package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	ActiveTunnels     prometheus.Gauge
	BytesSentTotal    prometheus.Counter
	BytesReceivedTotal prometheus.Counter
	ConnectionAttempts prometheus.Counter
	ErrorsTotal       prometheus.Counter

	registry *prometheus.Registry
}

// NewMetrics creates and registers all metrics
func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		ActiveTunnels: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "psiphond_ng_active_tunnels",
			Help: "Number of currently active tunnels",
		}),
		BytesSentTotal: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "psiphond_ng_bytes_sent_total",
			Help: "Total number of bytes sent through tunnels",
		}),
		BytesReceivedTotal: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "psiphond_ng_bytes_received_total",
			Help: "Total number of bytes received through tunnels",
		}),
		ConnectionAttempts: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "psiphond_ng_connection_attempts_total",
			Help: "Total number of connection attempts",
		}),
		ErrorsTotal: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "psiphond_ng_errors_total",
			Help: "Total number of errors encountered",
		}),
	}
	m.registry = registry
	return m
}

// Handler returns the HTTP handler for /metrics endpoint
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// GetActiveTunnels returns the current value of active tunnels gauge
func (m *Metrics) GetActiveTunnels() (float64, error) {
	return getGaugeValue(m.registry, "psiphond_ng_active_tunnels")
}

// GetBytesSentTotal returns total bytes sent
func (m *Metrics) GetBytesSentTotal() (float64, error) {
	return getCounterValue(m.registry, "psiphond_ng_bytes_sent_total")
}

// GetBytesReceivedTotal returns total bytes received
func (m *Metrics) GetBytesReceivedTotal() (float64, error) {
	return getCounterValue(m.registry, "psiphond_ng_bytes_received_total")
}

// GetConnectionAttempts returns total connection attempts
func (m *Metrics) GetConnectionAttempts() (float64, error) {
	return getCounterValue(m.registry, "psiphond_ng_connection_attempts_total")
}

// GetErrorsTotal returns total errors
func (m *Metrics) GetErrorsTotal() (float64, error) {
	return getCounterValue(m.registry, "psiphond_ng_errors_total")
}

// getGaugeValue retrieves a gauge value from the registry
func getGaugeValue(registry *prometheus.Registry, name string) (float64, error) {
	mfs, err := registry.Gather()
	if err != nil {
		return 0, err
	}
	for _, mf := range mfs {
		if mf.GetName() == name {
			if len(mf.GetMetric()) == 0 {
				return 0, nil
			}
			return mf.GetMetric()[0].GetGauge().GetValue(), nil
		}
	}
	return 0, fmt.Errorf("metric %s not found", name)
}

// getCounterValue retrieves a counter value from the registry
func getCounterValue(registry *prometheus.Registry, name string) (float64, error) {
	mfs, err := registry.Gather()
	if err != nil {
		return 0, err
	}
	for _, mf := range mfs {
		if mf.GetName() == name {
			if len(mf.GetMetric()) == 0 {
				return 0, nil
			}
			return mf.GetMetric()[0].GetCounter().GetValue(), nil
		}
	}
	return 0, fmt.Errorf("metric %s not found", name)
}
