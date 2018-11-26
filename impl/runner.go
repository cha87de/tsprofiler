package impl

import (
	"time"

	"github.com/cha87de/tsprofiler/spec"
)

func (profiler *profiler) profileOutputRunner() {
	if profiler.settings.OutputCallback == nil || profiler.settings.OutputFreq == 0 {
		// no automated output specified
		return
	}
	for !profiler.stopped {
		start := time.Now()
		profile := profiler.generateProfile()
		profiler.settings.OutputCallback(profile)
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

func (profiler *profiler) generateProfile() spec.TSProfile {
	var metrics []spec.TSProfileMetric
	profiler.metricsAccess.Lock()
	defer profiler.metricsAccess.Unlock()
	for _, metricProfiler := range profiler.metrics {
		maxCount := float64(metricProfiler.counts.stats.Count) / float64(profiler.settings.BufferSize)
		txmatrix := computeProbabilities(metricProfiler.counts.stateChangeCounter, maxCount)
		// fmt.Printf("counter %+v, probs: %+v\n", metricProfiler.counts.stateChangeCounter, txmatrix)
		metrics = append(metrics, spec.TSProfileMetric{
			Name:     metricProfiler.name,
			TXMatrix: txmatrix,
			Stats:    metricProfiler.counts.stats,
		})
	}
	return spec.TSProfile{
		Name:     profiler.settings.Name,
		Metrics:  metrics,
		Settings: profiler.settings,
	}
}
