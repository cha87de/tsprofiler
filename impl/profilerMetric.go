package impl

import (
	"fmt"
	"math"
	"sync"

	"github.com/cha87de/tsprofiler/spec"
	"gonum.org/v1/gonum/stat"
)

func newProfilerMetric(name string, maxstates int, filterStdDevs int) profilerMetric {
	metric := profilerMetric{
		name: name,
		buffer: &profilerMetricBuffer{
			rawData:       make([]float64, 0),
			access:        &sync.Mutex{},
			min:           -1,
			filterStdDevs: filterStdDevs,
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

func (profilerMetric *profilerMetric) isOutlier(value float64) bool {
	if profilerMetric.counts.stats.Avg == 0 || profilerMetric.counts.stats.Stddev == 0 {
		return false
	}
	diff := math.Abs(value - profilerMetric.counts.stats.Avg)
	return diff >= float64(profilerMetric.buffer.filterStdDevs)*profilerMetric.counts.stats.Stddev
}

// countBuffer takes values from `buffer`, and counts discrete states in `counts`
func (profilerMetric *profilerMetric) countBuffer() {
	rawData, min, max := profilerMetric.buffer.reset()
	bufferAverage := avg(rawData)

	newState := discretize(bufferAverage, profilerMetric.counts.maxstates, min, max)
	if newState.value < 0 || newState.value >= int64(profilerMetric.counts.maxstates) {
		fmt.Printf("no valid state found (i) for value %v\n", bufferAverage)
		// no state found
		return
	}

	// min/max changed? update tx matrix
	if profilerMetric.counts.stats.Min > min || profilerMetric.counts.stats.Max < max {
		profilerMetric.counts.changeDimension(min, max)
	}

	// count new state transition
	oldState := profilerMetric.counts.currentState
	profilerMetric.counts.stateChangeCounter[oldState.value][newState.value]++
	profilerMetric.counts.currentState.value = newState.value

	// update global stats
	oldAvg := profilerMetric.counts.stats.Avg
	profilerMetric.counts.stats.Avg = stat.Mean([]float64{profilerMetric.counts.stats.Avg, bufferAverage}, []float64{float64(profilerMetric.counts.stats.Count), float64(len(rawData))})
	profilerMetric.counts.stats.Count += int64(len(rawData))
	for _, v := range rawData {
		profilerMetric.counts.stats.StddevSum += (v - oldAvg) * (v - profilerMetric.counts.stats.Avg)
	}
	profilerMetric.counts.stats.Stddev = math.Sqrt(profilerMetric.counts.stats.StddevSum / float64(profilerMetric.counts.stats.Count))

}

type profilerMetricBuffer struct {
	rawData       []float64
	access        *sync.Mutex
	min           float64
	max           float64
	filterStdDevs int
}

func (buffer *profilerMetricBuffer) append(value float64) {
	buffer.access.Lock()
	buffer.rawData = append(buffer.rawData, value)
	if value > buffer.max {
		buffer.max = value
	}
	if buffer.min == -1 || value < buffer.min {
		buffer.min = value
	}
	buffer.access.Unlock()
}

func (buffer *profilerMetricBuffer) reset() ([]float64, float64, float64) {
	rawData := make([]float64, len(buffer.rawData))
	buffer.access.Lock()
	copy(rawData, buffer.rawData)
	buffer.rawData = make([]float64, 0)
	buffer.access.Unlock()
	return rawData, buffer.min, buffer.max
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
	counts.stats = spec.TSStats{
		Min:       -1,
		Max:       -1,
		Avg:       0.0,
		Stddev:    0.0,
		Count:     0.0,
		StddevSum: 0.0,
	}
}

// changeDimension recomputes the state counter values for the new min/max dimension
func (counts *profilerMetricCounts) changeDimension(min float64, max float64) {
	newCounterMatrix := make([][]int64, counts.maxstates)
	for i := range newCounterMatrix {
		newCounterMatrix[i] = make([]int64, counts.maxstates)
	}

	oldMin := counts.stats.Min
	oldMax := counts.stats.Max
	maxstate := counts.maxstates
	oldStateStepSize := float64(oldMax-oldMin) / float64(maxstate)

	for i := range counts.stateChangeCounter {
		valueI := float64(0)
		newStateI := state{
			value: -1,
		}
		for j := range counts.stateChangeCounter[i] {
			oldCounter := counts.stateChangeCounter[i][j]
			// were there any occurrences at all?
			if oldCounter <= 0 {
				continue
			}

			if newStateI.value == -1 {
				// lazy compute: state for i not yet calculated
				valueI = float64(i) * oldStateStepSize
				valueI += oldMin
				newStateI = discretize(valueI, maxstate, min, max)
			}
			valueJ := float64(j) * oldStateStepSize
			valueJ += oldMin
			newStateJ := discretize(valueJ, maxstate, min, max)

			if newStateI.value < 0 || newStateI.value >= int64(maxstate) {
				fmt.Printf("no valid state found (iI). %.0f + %.0f * %d = %.0f (min %v, max %v, oldmin %v, oldmax %v)\n", oldMin, oldStateStepSize, i, valueI, min, max, oldMin, oldMax)
				// no state found
				continue
			}
			if newStateJ.value < 0 || newStateJ.value >= int64(maxstate) {
				fmt.Printf("no valid state found (iJ) for value %v (min: %v, max %v, j: %v, stepsize: %v)\n", valueJ, min, max, j, oldStateStepSize)
				// no state found
				continue
			}
			//fmt.Printf("%+v,%+v\n", newStateI.value, newStateJ.value)
			newCounterMatrix[newStateI.value][newStateJ.value] += oldCounter
		}
	}
	counts.stateChangeCounter = newCounterMatrix

	if counts.stats.Min == -1 || counts.stats.Min > min {
		counts.stats.Min = min
	}
	if counts.stats.Max < max {
		counts.stats.Max = max
	}
}
