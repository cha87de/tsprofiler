package spec

import (
	"time"
)

// TSProfiler defines the profilers interface to transfer time series data into
// a statistical profile
type TSProfiler interface {
	// Put allows applications to provide a new TSData input to the profiler
	Put(data TSData)
}

// Settings defines settings for TSProfiler
type Settings struct {
	// BufferSize defines the amount of TSData items before a new state is transitioned
	BufferSize int

	// Name allows to identify the profiler, e.g. for human readable differentiation
	Name string

	// States defines the amount of states to discretize the measurements
	States int

	// OutputFreq controls the frequency in which the profiler calls the OutputCallback function
	OutputFreq time.Duration

	// OutputCallback defines the callback function for `TSProfile`s every `OutputFreq`
	OutputCallback func(data TSProfile)
}
