package spec

// TSProfile contains the resulting statistical profile
type TSProfile struct {
	Name    string            `json:"name"`
	Metrics []TSProfileMetric `json:"metrics"`
}

// TSProfileMetric ...
type TSProfileMetric struct {
	Name     string  `json:"name"`
	TXMatrix [][]int `json:"txmatrix"`
}
