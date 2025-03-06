package restful_finder

import (
	"time"
)

type Trace interface {
	RecordErr(label, key string, err error, tag string)
	PushWaitingTaskEnd(label, key string, cost time.Duration)
	RecordAPIEnd(label, key string, cost time.Duration)
	ScanRestfulPatternEnd(cost time.Duration, normalList []URLWithLabel, restfulList []URLWithLabel)
	ActiveWaitingTaskEnd(cost time.Duration, count int)
	ClearEnd(cost time.Duration)
}

type MetricTrace struct {
}

func (m *MetricTrace) RecordErr(label, key string, err error, tag string) {
	ERRCounter.WithLabelValues(label, key, err.Error(), tag).Inc()
}

func (m *MetricTrace) PushWaitingTaskEnd(label, key string, cost time.Duration) {
	ActionHistory.WithLabelValues(label, key, "push_waiting_task").Observe(cost.Seconds())
}

func (m *MetricTrace) RecordAPIEnd(label, key string, cost time.Duration) {
	ActionHistory.WithLabelValues(label, key, "record_api").Observe(cost.Seconds())
}

func (m *MetricTrace) ScanRestfulPatternEnd(cost time.Duration, normalList []URLWithLabel, restfulList []URLWithLabel) {
	ActionHistory.WithLabelValues("", "", "scan_restful_pattern").Observe(cost.Seconds())

	for _, item := range normalList {
		APICounter.WithLabelValues(item.Label, item.URL, "否").Inc()
	}

	for _, item := range restfulList {
		APICounter.WithLabelValues(item.Label, item.URL, "是").Inc()
	}
}

func (m *MetricTrace) ActiveWaitingTaskEnd(cost time.Duration, count int) {
	ActionHistory.WithLabelValues("", "", "active_waiting_task").Observe(cost.Seconds())
	ActiveTaskCounter.WithLabelValues().Add(float64(count))
}

func (m *MetricTrace) ClearEnd(cost time.Duration) {
	ActionHistory.WithLabelValues("", "", "clear").Observe(cost.Seconds())
}



