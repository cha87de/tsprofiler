package spec

// TSProfiler defines the profilers interface to transfer time series data into
// a statistical profile
type TSProfiler interface {
	// Initialize(settings Settings) error
	InputPipe() chan TSData
	Put(data TSData)
}

// Settings defines settings for TSProfiler
type Settings struct {
	Frequency int
	Name      string
}
