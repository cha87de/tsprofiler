package models

// NewTSBuffer instantiates a new TSBuffer with given metric name
func NewTSBuffer(metric string) TSBuffer {
	return TSBuffer{
		Metric:  metric,
		RawData: make([]float64, 0),
		Min:     -1,
	}
}

// TSBuffer describes one full buffer
type TSBuffer struct {
	Metric   string
	RawData  []float64
	Min      float64
	Max      float64
	FixedMin float64
	FixedMax float64
}

// Append adds a single value to the TSBuffer
func (tsbuffer *TSBuffer) Append(value float64) {
	tsbuffer.RawData = append(tsbuffer.RawData, value)
	// dynamic min/max ranges
	if value > tsbuffer.Max {
		tsbuffer.Max = value
	}
	if tsbuffer.Min == -1 || value < tsbuffer.Min {
		tsbuffer.Min = value
	}
}
