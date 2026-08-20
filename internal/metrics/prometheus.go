package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metric collectors for the application.
var (
	// HTTPRequestsTotal counts total HTTP requests by method, path, and status code.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ratelimiterx",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests processed.",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration observes request latency in seconds by method and path.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ratelimiterx",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// RateLimitHitsTotal counts requests that were rate-limited (429) by tier and algorithm.
	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ratelimiterx",
			Name:      "rate_limit_hits_total",
			Help:      "Total number of requests that were rate-limited.",
		},
		[]string{"tier", "algorithm"},
	)

	// RateLimitAllowedTotal counts requests that passed rate limiting by tier and algorithm.
	RateLimitAllowedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ratelimiterx",
			Name:      "rate_limit_allowed_total",
			Help:      "Total number of requests that passed rate limiting.",
		},
		[]string{"tier", "algorithm"},
	)

	// ActiveConnections tracks the current number of active HTTP connections.
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ratelimiterx",
			Name:      "active_connections",
			Help:      "Number of currently active HTTP connections.",
		},
	)

	// RedisOperationsTotal counts Redis operations by operation type and result.
	RedisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ratelimiterx",
			Name:      "redis_operations_total",
			Help:      "Total number of Redis operations.",
		},
		[]string{"operation", "result"},
	)

	// RedisOperationDuration observes Redis operation latency.
	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ratelimiterx",
			Name:      "redis_operation_duration_seconds",
			Help:      "Redis operation latency in seconds.",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25},
		},
		[]string{"operation"},
	)
)
