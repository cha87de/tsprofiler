package impl

import (
	"time"

	"github.com/cha87de/tsprofiler/spec"
)

func (profiler *simpleProfiler) profileOutputRunner() {
	for {
		start := time.Now()
		profiler.output()
		nextRun := start.Add(time.Duration(20) * time.Second)
		time.Sleep(nextRun.Sub(time.Now()))
	}
}

func (profiler *simpleProfiler) profile() {
	// only considering CPU at the moment!
	// TODO add IO and Net
	profiler.dataaccess.Lock()
	var cpudata []float64
	for _, d := range profiler.cpudata {
		cpudata = append(cpudata, d.CPU)
	}
	profiler.cpudata = make([]spec.TSData, 0)
	profiler.dataaccess.Unlock()

	newCPUState := discretize(aggregate(cpudata))
	profiler.transit(newCPUState)
}

func (profiler *simpleProfiler) transit(state state) {
	// only considering CPU at the moment!
	// TODO add IO and Net
	profiler.cpu.statematrix[profiler.cpu.currentState.value][state.value]++
	// finally: update current state
	profiler.cpu.currentState = state
}

func (profiler *simpleProfiler) output() {
	cpuProb := computeProbabilities(profiler.cpu.statematrix)
	ioProb := computeProbabilities(profiler.io.statematrix)
	netProb := computeProbabilities(profiler.net.statematrix)

	var metrics []spec.TSProfileMetric
	metrics = append(metrics, spec.TSProfileMetric{
		Name:     "cpu",
		TXMatrix: cpuProb,
	})
	metrics = append(metrics, spec.TSProfileMetric{
		Name:     "io",
		TXMatrix: ioProb,
	})
	metrics = append(metrics, spec.TSProfileMetric{
		Name:     "net",
		TXMatrix: netProb,
	})

	profiler.settings.OutputCallback(spec.TSProfile{
		Name:    profiler.settings.Name,
		Metrics: metrics,
	})
}
