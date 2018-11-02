package impl

import (
	"time"

	"github.com/cha87de/tsprofiler/spec"
)

func (profiler *profiler) profileOutputRunner() {
	for !profiler.stopped {
		start := time.Now()
		profiler.output()
		nextRun := start.Add(profiler.settings.OutputFreq)
		time.Sleep(nextRun.Sub(time.Now()))
	}
}

// take profilers
func (profiler *profiler) profile() {
	profiler.metricsAccess.Lock()
	for _, metricProfiler := range profiler.metrics {
		metricProfiler.countBuffer()
	}
	profiler.metricsAccess.Unlock()
}

func (profiler *profiler) output() {
	var metrics []spec.TSProfileMetric
	profiler.metricsAccess.Lock()
	for _, metricProfiler := range profiler.metrics {
		txmatrix := computeProbabilities(metricProfiler.counts.stateChangeCounter)
		//fmt.Printf("counter %+v, probs: %+v\n", metricProfiler.counts.stateChangeCounter, txmatrix)
		metrics = append(metrics, spec.TSProfileMetric{
			Name:     metricProfiler.name,
			TXMatrix: txmatrix,
		})
	}
	profiler.metricsAccess.Unlock()
	profiler.settings.OutputCallback(spec.TSProfile{
		Name:    profiler.settings.Name,
		Metrics: metrics,
	})
}
