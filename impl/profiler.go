package impl

import (
	"sync"

	"github.com/cha87de/tsprofiler/spec"
)

// NewSimpleProfiler creates and returns a new TSProfiler, configured with given Settings
func NewSimpleProfiler(settings spec.Settings) *simpleProfiler {
	profiler := simpleProfiler{}
	profiler.initialize(settings)
	return &profiler
}

// SimpleProfiler implements a simple aggregation based TSProfiler
type simpleProfiler struct {
	in       chan spec.TSData
	settings spec.Settings

	data       []spec.TSData
	dataaccess *sync.Mutex

	currentState state
	statematrix  [][]int64
}

func (profiler *simpleProfiler) initialize(settings spec.Settings) error {
	profiler.settings = settings
	profiler.data = make([]spec.TSData, 0)
	profiler.dataaccess = &sync.Mutex{}
	profiler.in = make(chan spec.TSData, 10)

	// initialize state matrix
	profiler.statematrix = make([][]int64, maxstates)
	for i := range profiler.statematrix {
		profiler.statematrix[i] = make([]int64, maxstates)
	}

	go profiler.profileRunner()
	go profiler.profilePrintRunner()
	go profiler.listener()
	return nil
}

func (profiler *simpleProfiler) InputPipe() chan spec.TSData {
	return profiler.in
}

func (profiler *simpleProfiler) Put(data spec.TSData) {
	profiler.in <- data
}
