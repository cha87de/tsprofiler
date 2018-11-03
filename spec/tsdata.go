package spec

// TSData describes a ts data point used as profiler input with a metrics array
type TSData struct {
	Metrics []TSDataMetric
}

// TSDataMetric describes profiler input for a single metrix
type TSDataMetric struct {
	Name  string
	Value float64
	Max   float64
}
