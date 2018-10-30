package impl

import (
	"sync"
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
	rawData := profilerMetric.buffer.reset()
	avg := avg(rawData)
	//min := min(rawData)
	//max := max(rawData)
	//stddev := stddev(rawData)
	min := float64(0)
	max := float64(100)
	newState := discretize(avg, profilerMetric.counts.maxstates, min, max)

	oldState := profilerMetric.counts.currentState
	// fmt.Printf("value: %+v, oldState: %+v, newState: %+v\n", aggregatedValue, newState, oldState)
	profilerMetric.counts.stateChangeCounter[oldState.value][newState.value]++
	profilerMetric.counts.currentState.value = newState.value
}

type profilerMetricBuffer struct {
	rawData []float64
	access  *sync.Mutex
}

func (buffer *profilerMetricBuffer) append(value float64) {
	buffer.access.Lock()
	buffer.rawData = append(buffer.rawData, value)
	buffer.access.Unlock()
}

func (buffer *profilerMetricBuffer) reset() []float64 {
	rawData := make([]float64, len(buffer.rawData))
	buffer.access.Lock()
	copy(rawData, buffer.rawData)
	buffer.rawData = make([]float64, 0)
	buffer.access.Unlock()
	return rawData
}

type profilerMetricCounts struct {
	maxstates          int
	currentState       state
	stateChangeCounter [][]int64
}

func (counts *profilerMetricCounts) reset() {
	counts.stateChangeCounter = make([][]int64, counts.maxstates)
	for i := range counts.stateChangeCounter {
		counts.stateChangeCounter[i] = make([]int64, counts.maxstates)
	}
	counts.currentState = state{
		value: 0,
	}
}
