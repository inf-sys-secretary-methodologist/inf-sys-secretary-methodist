// Package metrics provides Prometheus metrics for monitoring and observability.
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTPRequestsTotal counts total HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration measures HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestsInFlight tracks current in-flight requests
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// DatabaseQueriesTotal counts total database queries
	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	// DatabaseQueryDuration measures database query duration
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "table"},
	)

	// CacheOperationsTotal counts cache operations
	CacheOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "result"},
	)

	// AuthEventsTotal counts authentication events
	AuthEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_events_total",
			Help: "Total number of authentication events",
		},
		[]string{"event", "status"},
	)

	// BusinessOperationsTotal counts business operations
	BusinessOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "business_operations_total",
			Help: "Total number of business operations",
		},
		[]string{"module", "operation", "status"},
	)

	// ActiveConnections tracks active connections
	ActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
		[]string{"type"},
	)
)

// PrometheusMiddleware returns a Gin middleware that collects Prometheus metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint to avoid recursion
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		HTTPRequestsInFlight.Inc()

		c.Next()

		HTTPRequestsInFlight.Dec()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := normalizePath(c.FullPath(), c.Request.URL.Path)

		HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		HTTPRequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
	}
}

// Handler returns the Prometheus metrics handler for Gin
func Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// normalizePath normalizes the path for metrics to avoid high cardinality
func normalizePath(fullPath, requestPath string) string {
	// Use the route pattern if available (e.g., /users/:id instead of /users/123)
	if fullPath != "" {
		return fullPath
	}
	// Fallback to request path for unmatched routes
	return requestPath
}

// RecordDatabaseQuery records a database query metric
func RecordDatabaseQuery(operation, table, status string, duration time.Duration) {
	DatabaseQueriesTotal.WithLabelValues(operation, table, status).Inc()
	DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordCacheOperation records a cache operation metric
func RecordCacheOperation(operation string, hit bool) {
	result := "miss"
	if hit {
		result = "hit"
	}
	CacheOperationsTotal.WithLabelValues(operation, result).Inc()
}

// RecordAuthEvent records an authentication event metric
func RecordAuthEvent(event string, success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	AuthEventsTotal.WithLabelValues(event, status).Inc()
}

// RecordBusinessOperation records a business operation metric
func RecordBusinessOperation(module, operation string, success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	BusinessOperationsTotal.WithLabelValues(module, operation, status).Inc()
}

// SetActiveConnections sets the number of active connections
func SetActiveConnections(connType string, count float64) {
	ActiveConnections.WithLabelValues(connType).Set(count)
}
