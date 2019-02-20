package models

// TSInputMetric describes profiler input for a single metric
type TSInputMetric struct {
	Name     string
	Value    float64
	FixedMin float64
	FixedMax float64
}
