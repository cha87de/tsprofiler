package spec

// TSProfile contains the resulting statistical profile
type TSProfile struct {
	Name    string            `json:"name"`
	Metrics []TSProfileMetric `json:"metrics"`
}

// TSStats contains default statistics
type TSStats struct {
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Stddev float64 `json:"stddev"`
	Avg    float64 `json:"avg"`
}

// TSProfileMetric describes for one metric a statistical profile
type TSProfileMetric struct {
	Name     string  `json:"name"`
	TXMatrix [][]int `json:"txmatrix"`
	Stats    TSStats `json:"stats"`
}
