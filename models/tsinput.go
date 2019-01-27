package models

// TSInput describes a ts data point used as profiler input with a metrics array
type TSInput struct {
	Metrics []TSInputMetric
}

// GetMetrics returns a list of metrics contained in TSInput
func (tsinput *TSInput) GetMetrics() []string {
	metrics := make([]string, len(tsinput.Metrics))
	for i, m := range tsinput.Metrics {
		metrics[i] = m.Name
	}
	return metrics
}

// Get returns the TSInputMetric for the given metric name
func (tsinput *TSInput) Get(metricname string) TSInputMetric {
	for _, m := range tsinput.Metrics {
		if m.Name == metricname {
			return m
		}
	}
	return TSInputMetric{}
}
