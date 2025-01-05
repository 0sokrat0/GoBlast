package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

var (
	TaskCreatedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_created_total",
			Help: "Общее количество созданных задач",
		},
	)

	TaskFailedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_failed_total",
			Help: "Количество неуспешных задач",
		},
	)

	TaskProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "task_processing_duration_seconds",
			Help:    "Длительность обработки задач",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"task_type"}, // Метка для группировки метрик по типу задачи
	)

	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Количество HTTP запросов",
		},
		[]string{"method", "endpoint"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность HTTP запросов",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Use sync.Once to ensure metrics are registered only once.
	registerOnce sync.Once
)

// InitMetrics регистрирует метрики в Prometheus
func InitMetrics() {
	registerOnce.Do(func() {
		prometheus.MustRegister(TaskCreatedCounter)
		prometheus.MustRegister(TaskFailedCounter)
		prometheus.MustRegister(TaskProcessingDuration)
		prometheus.MustRegister(RequestCounter)
		prometheus.MustRegister(RequestDuration)
	})
}

// RegisterMetricsEndpoint регистрирует endpoint для метрик
func RegisterMetricsEndpoint() {
	http.Handle("/metrics", promhttp.Handler())
}
