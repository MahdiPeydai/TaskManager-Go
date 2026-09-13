package metrics

import "github.com/prometheus/client_golang/prometheus"

var DbCall = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "db_calls_total",
		Help: "Number of database calls",
	},
	[]string{"model", "operation", "status"},
)

var RequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	},
	[]string{"method", "path", "status_code"},
)
