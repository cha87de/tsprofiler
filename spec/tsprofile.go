package spec

// TSProfile contains the resulting statistical profile
type TSProfile struct {
	Name     string            `json:"name"`
	Metrics  []TSProfileMetric `json:"metrics"`
	Settings Settings          `json:"settings"`
}

// TSStats contains default statistics
type TSStats struct {
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Stddev    float64 `json:"stddev"`
	Avg       float64 `json:"avg"`
	Count     int64   `json:"count"`
	StddevSum float64 `json:"-"`
}

// TSProfileMetric describes for one metric a statistical profile
type TSProfileMetric struct {
	Name     string            `json:"name"`
	TXMatrix map[string]TXStep `json:"txmatrix"`
	Stats    TSStats           `json:"stats"`
}

// TXStep expresses a single state in a markov chain / transition matrix
type TXStep struct {
	NextStateProbs []int `json:"nextProbs"`
	StepProb       int   `json:"probability"`
}
