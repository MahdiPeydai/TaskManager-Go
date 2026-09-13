package metrics

import "github.com/prometheus/client_golang/prometheus"

var TaskCount = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "task_count",
		Help: "Current number of tasks.",
	},
)
