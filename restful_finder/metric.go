package restful_finder

import "github.com/prometheus/client_golang/prometheus"

var (
	ERRCounter        *prometheus.CounterVec
	ActionHistory     *prometheus.HistogramVec
	APICounter        *prometheus.CounterVec
	ActiveTaskCounter *prometheus.CounterVec
)

func GetMetricCollectors(ns string) []prometheus.Collector {
	subSystem := "restful_finder"
	ERRCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: subSystem,
			Name:      "error",
			Help:      "error detail",
		},
		[]string{"label", "key", "tag"},
	)

	ActionHistory = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: subSystem,
			Name:      "action",
			Help:      "the cost and qps when we do action like record api...",
		},
		[]string{"action", "label", "key"},
	)

	APICounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: subSystem,
			Name:      "api",
			Help:      "after scan , the api list we get",
		},
		[]string{"label", "api", "is_restful"},
	)

	ActiveTaskCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: subSystem,
			Name:      "active_task",
			Help:      "active task count",
		},
		[]string{},
	)

	return []prometheus.Collector{
		ERRCounter,
		ActionHistory,
		APICounter,
		ActiveTaskCounter,
	}
}
