package models

// TSInputMetric describes profiler input for a single metrix
type TSInputMetric struct {
	Name     string
	Value    float64
	FixedMin float64
	FixedMax float64
}
