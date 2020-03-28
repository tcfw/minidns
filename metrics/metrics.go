package metrics

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//Store holds various prom metrics
type Store struct {
	requests      *prometheus.CounterVec
	pluginMetrics map[string]prometheus.Metric
}

var (
	metrics = newMetrics()
)

func newMetrics() *Store {
	return &Store{
		requests: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "minidns_request_totals",
			Help: "Total number of requests processed",
		}, []string{"type"}),
		pluginMetrics: map[string]prometheus.Metric{},
	}
}

//RegisterHTTPHandler register the promhttp handler on /metrics
func RegisterHTTPHandler() {
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Register Prometheus metrics endpoint")
}

//IncRequests increment the requests request counter
func (m *Store) IncRequests(label string) {
	m.requests.WithLabelValues(label).Inc()
}

//RegisterPluginMetric add a custom metric
func (m *Store) RegisterPluginMetric(name string, metric prometheus.Metric) error {
	if _, ok := m.pluginMetrics[name]; ok {
		return fmt.Errorf("metric name already registered")
	}

	m.pluginMetrics[name] = metric
	return nil
}

//GetPMetric get a plugin metric
func GetPMetric(name string) prometheus.Metric {
	if metric, ok := metrics.pluginMetrics[name]; ok {
		return metric
	}

	return nil
}

//GetMetrics metrics instance
func GetMetrics() *Store {
	return metrics
}

//IncRequests increment the handled request counter
func IncRequests(label string) {
	metrics.IncRequests(label)
}
