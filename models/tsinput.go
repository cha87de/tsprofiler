package models

// TSInput describes a ts data point used as profiler input with a metrics array
type TSInput struct {
	Metrics []TSInputMetric `json:"metrics"`
}
