package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all application metrics
type Metrics struct {
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestsInFlight  prometheus.Gauge
	databaseConnections   prometheus.Gauge
	databaseQueries       *prometheus.CounterVec
	databaseQueryDuration *prometheus.HistogramVec
	authAttempts          *prometheus.CounterVec
}

// NewMetrics creates and registers all metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		httpRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),
		databaseConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "database_connections_active",
				Help: "Number of active database connections",
			},
		),
		databaseQueries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),
		databaseQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
		authAttempts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_attempts_total",
				Help: "Total number of authentication attempts",
			},
			[]string{"type", "status"},
		),
	}

	// Register all metrics
	prometheus.MustRegister(
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestsInFlight,
		m.databaseConnections,
		m.databaseQueries,
		m.databaseQueryDuration,
		m.authAttempts,
	)

	return m
}

// MetricsMiddleware collects HTTP metrics
func (m *Metrics) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment in-flight requests
		m.httpRequestsInFlight.Inc()
		defer m.httpRequestsInFlight.Dec()

		// Process request
		c.Next()

		// Collect metrics
		duration := time.Since(start).Seconds()
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		m.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		m.httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)
	}
}

// RecordDatabaseQuery records database query metrics
func (m *Metrics) RecordDatabaseQuery(operation, table string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	m.databaseQueries.WithLabelValues(operation, table, status).Inc()
	m.databaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordAuthAttempt records authentication attempt metrics
func (m *Metrics) RecordAuthAttempt(authType string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}

	m.authAttempts.WithLabelValues(authType, status).Inc()
}

// SetDatabaseConnections sets the number of active database connections
func (m *Metrics) SetDatabaseConnections(count float64) {
	m.databaseConnections.Set(count)
}

// Handler returns the Prometheus metrics HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// IncrementCounter provides a generic counter increment method
func (m *Metrics) IncrementCounter(name string) {
	switch name {
	case "order_processing_failures", "order_processing_success":
		// For now, just use a generic counter or extend the metrics struct
		// This is a simplified implementation
		m.httpRequestsTotal.WithLabelValues("POST", "/orders", "200").Inc()
	case "order_refund_failures", "order_refund_success":
		m.httpRequestsTotal.WithLabelValues("POST", "/orders/refund", "200").Inc()
	}
}

// HealthMetrics provides basic health metrics
type HealthMetrics struct {
	StartTime    time.Time
	DatabaseUp   bool
	ExternalAPIs map[string]bool
}

// NewHealthMetrics creates a new health metrics instance
func NewHealthMetrics() *HealthMetrics {
	return &HealthMetrics{
		StartTime:    time.Now(),
		ExternalAPIs: make(map[string]bool),
	}
}

// Uptime returns the application uptime
func (h *HealthMetrics) Uptime() time.Duration {
	return time.Since(h.StartTime)
}

// SetDatabaseStatus sets the database health status
func (h *HealthMetrics) SetDatabaseStatus(up bool) {
	h.DatabaseUp = up
}

// SetExternalAPIStatus sets the status of an external API
func (h *HealthMetrics) SetExternalAPIStatus(name string, up bool) {
	h.ExternalAPIs[name] = up
}

// IsHealthy returns true if all systems are healthy
func (h *HealthMetrics) IsHealthy() bool {
	if !h.DatabaseUp {
		return false
	}

	for _, status := range h.ExternalAPIs {
		if !status {
			return false
		}
	}

	return true
}
