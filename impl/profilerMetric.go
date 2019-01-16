package impl

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/cha87de/tsprofiler/spec"
	"gonum.org/v1/gonum/stat"
)

func newProfilerMetric(name string, maxstates int, history int, filterStdDevs int, fixbound bool) profilerMetric {
	metric := profilerMetric{
		name: name,
		buffer: &profilerMetricBuffer{
			rawData:       make([]float64, 0),
			access:        &sync.Mutex{},
			min:           -1,
			filterStdDevs: filterStdDevs,
			fixbound:      fixbound,
		},
		counts: &profilerMetricCounts{
			maxstates: maxstates,
			history:   history,
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

	for len(oldState) > 0 {
		oldStateIdent := ""
		for _, state := range oldState {
			if oldStateIdent != "" {
				oldStateIdent = oldStateIdent + "-"
			}
			oldStateIdent = oldStateIdent + fmt.Sprintf("%d", state.value)
		}
		_, ok := profilerMetric.counts.stateChangeCounter[oldStateIdent]
		if !ok {
			profilerMetric.counts.stateChangeCounter[oldStateIdent] = make([]int64, profilerMetric.counts.maxstates)
		}
		profilerMetric.counts.stateChangeCounter[oldStateIdent][newState.value]++
		oldState = oldState[1:]
	}

	if len(profilerMetric.counts.currentState) > 0 {
		profilerMetric.counts.currentState = profilerMetric.counts.currentState[1:] // remove first item
	}
	profilerMetric.counts.currentState = append(profilerMetric.counts.currentState, newState) // add new item at the end

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
	fixbound      bool
}

func (buffer *profilerMetricBuffer) append(value spec.TSDataMetric) {
	buffer.access.Lock()
	buffer.rawData = append(buffer.rawData, value.Value)
	if buffer.fixbound {
		// use fix min/max ranges
		buffer.min = value.FixedMin
		buffer.max = value.FixedMax
	} else {
		// dynamic min/max ranges
		if value.Value > buffer.max {
			buffer.max = value.Value
		}
		if buffer.min == -1 || value.Value < buffer.min {
			buffer.min = value.Value
		}
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
	history            int
	currentState       []state
	stateChangeCounter map[string][]int64
	stats              spec.TSStats
}

func (counts *profilerMetricCounts) reset() {
	counts.stateChangeCounter = make(map[string][]int64)
	counts.currentState = make([]state, counts.history)
	for i := range counts.currentState {
		counts.currentState[i] = state{
			value: 0,
		}
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
	newCounterMatrix := make(map[string][]int64)

	oldMin := counts.stats.Min
	oldMax := counts.stats.Max
	maxstate := counts.maxstates
	oldStateStepSize := float64(oldMax-oldMin) / float64(maxstate)

	for key := range counts.stateChangeCounter {
		var newKey string
		for j := range counts.stateChangeCounter[key] {
			oldCounter := counts.stateChangeCounter[key][j]
			// were there any occurrences at all?
			if oldCounter <= 0 {
				continue
			}

			if newKey == "" {
				// lazy compute: state for i not yet calculated
				keyParts := strings.Split(key, "-")
				for _, keyPart := range keyParts {
					i, err := strconv.ParseInt(keyPart, 10, 32)
					if err != nil {
						i = 0
					}
					valueIpart := float64(i) * oldStateStepSize
					valueIpart += oldMin
					newStateIpart := discretize(valueIpart, maxstate, min, max)
					if newStateIpart.value < 0 || newStateIpart.value >= int64(maxstate) {
						fmt.Printf("no valid state found (iI). %.0f + %.0f * %s = %.0f (min %v, max %v, oldmin %v, oldmax %v)\n", oldMin, oldStateStepSize, key, valueIpart, min, max, oldMin, oldMax)
						// no state found
						newKey = ""
						break
					}
					if newKey != "" {
						newKey = newKey + "-"
					}
					newKey = newKey + fmt.Sprintf("%d", newStateIpart.value)
				}
			}
			if newKey == "" {
				// if still empty, we have invalid states
				continue
			}
			valueJ := float64(j) * oldStateStepSize
			valueJ += oldMin
			newStateJ := discretize(valueJ, maxstate, min, max)

			if newStateJ.value < 0 || newStateJ.value >= int64(maxstate) {
				fmt.Printf("no valid state found (iJ) for value %v (min: %v, max %v, j: %v, stepsize: %v)\n", valueJ, min, max, j, oldStateStepSize)
				// no state found
				continue
			}
			//fmt.Printf("%+v,%+v\n", newStateI.value, newStateJ.value)
			_, ok := newCounterMatrix[newKey]
			if !ok {
				newCounterMatrix[newKey] = make([]int64, counts.maxstates)
			}
			newCounterMatrix[newKey][newStateJ.value] += oldCounter
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
