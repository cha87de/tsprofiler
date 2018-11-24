package impl

import (
	"sync"

	"github.com/cha87de/tsprofiler/spec"
)

// NewProfiler creates and returns a new TSProfiler, configured with given Settings
func NewProfiler(settings spec.Settings) *profiler {
	profiler := profiler{}
	profiler.initialize(settings)
	return &profiler
}

// profiler implements a  aggregation based TSProfiler
type profiler struct {
	input    chan spec.TSData
	settings spec.Settings

	metrics       []profilerMetric
	metricsAccess *sync.Mutex

	stopped bool
}

func (profiler *profiler) initialize(settings spec.Settings) {
	profiler.input = make(chan spec.TSData, 0)
	profiler.settings = settings
	profiler.metricsAccess = &sync.Mutex{}
	profiler.stopped = false
	go profiler.profileOutputRunner()
	go profiler.listener()
}

// Put adds a TSData item to the profiler
func (profiler *profiler) Put(data spec.TSData) {
	profiler.input <- data
}

// Get generates an returns a profile based on previously put data
func (profiler *profiler) Get() spec.TSProfile {
	return profiler.generateProfile()
}

// Terminate stops and removes the profiler
func (profiler *profiler) Terminate() {
	profiler.stopped = true
	close(profiler.input)
}

func (profiler *profiler) getMetricProfiler(name string) *profilerMetric {
	profiler.metricsAccess.Lock()

	// exists already?
	for _, metricProfiler := range profiler.metrics {
		if metricProfiler.name == name {
			profiler.metricsAccess.Unlock()
			return &metricProfiler
		}
	}

	// still here? create the profilerMetric
	metricProfiler := newProfilerMetric(name, profiler.settings.States, profiler.settings.History, profiler.settings.FilterStdDevs)
	profiler.metrics = append(profiler.metrics, metricProfiler)

	profiler.metricsAccess.Unlock()
	return &metricProfiler
}

func (profiler *profiler) add(data spec.TSData) {
	for _, metric := range data.Metrics {
		metricProfiler := profiler.getMetricProfiler(metric.Name)
		isOutlier := metricProfiler.isOutlier(metric.Value)
		if !isOutlier {
			metricProfiler.buffer.append(metric.Value)
			// } else {
			// fmt.Printf("skipped outlier %0.f\n", metric.Value)
		}
	}
}
