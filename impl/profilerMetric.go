package impl

import (
	"sync"

	"github.com/cha87de/tsprofiler/spec"
)

func newProfilerMetric(name string, maxstates int) profilerMetric {
	metric := profilerMetric{
		name: name,
		buffer: &profilerMetricBuffer{
			rawData: make([]float64, 0),
			access:  &sync.Mutex{},
		},
		counts: &profilerMetricCounts{
			maxstates: maxstates,
			// rest will be initialized via counts.reset()
		},
	}
	metric.counts.reset()
	return metric
}

type profilerMetric struct {
	name   string
	buffer *profilerMetricBuffer
	counts *profilerMetricCounts
}

// countBuffer takes values from `buffer`, and counts discrete states in `counts`
func (profilerMetric *profilerMetric) countBuffer() {
	rawData, max := profilerMetric.buffer.reset()
	avg := avg(rawData)
	//min := min(rawData)
	stddev := stddev(rawData)
	min := float64(0)
	newState := discretize(avg, profilerMetric.counts.maxstates, min, max)

	if newState.value < 0 {
		// no state found
		return
	}

	oldState := profilerMetric.counts.currentState
	// fmt.Printf("value: %+v, oldState: %+v, newState: %+v\n", avg, oldState, newState)
	profilerMetric.counts.stateChangeCounter[oldState.value][newState.value]++
	profilerMetric.counts.currentState.value = newState.value
	profilerMetric.counts.stats.Avg = avg
	profilerMetric.counts.stats.Min = min
	profilerMetric.counts.stats.Max = max
	profilerMetric.counts.stats.Stddev = stddev
}

type profilerMetricBuffer struct {
	rawData []float64
	access  *sync.Mutex
	max     float64
}

func (buffer *profilerMetricBuffer) append(value float64, max float64) {
	buffer.access.Lock()
	buffer.rawData = append(buffer.rawData, value)
	buffer.max = max
	buffer.access.Unlock()
}

func (buffer *profilerMetricBuffer) reset() ([]float64, float64) {
	rawData := make([]float64, len(buffer.rawData))
	buffer.access.Lock()
	copy(rawData, buffer.rawData)
	buffer.rawData = make([]float64, 0)
	buffer.access.Unlock()
	return rawData, buffer.max
}

type profilerMetricCounts struct {
	maxstates          int
	currentState       state
	stateChangeCounter [][]int64
	stats              spec.TSStats
}

func (counts *profilerMetricCounts) reset() {
	counts.stateChangeCounter = make([][]int64, counts.maxstates)
	for i := range counts.stateChangeCounter {
		counts.stateChangeCounter[i] = make([]int64, counts.maxstates)
	}
	counts.currentState = state{
		value: 0,
	}
	counts.stats = spec.TSStats{}
}
