package counter

import (
	"fmt"
	"math"
	"sync"

	"github.com/cha87de/tsprofiler/api"
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/utils"
	"gonum.org/v1/gonum/stat"
)

// NewCounter initializes and returns a new Counter
func NewCounter(history int, states int, buffersize int, profiler api.TSProfiler) Counter {
	return Counter{
		profiler: profiler,

		currentState:        make(map[string][]models.State),
		stateChangeCounters: make(map[string]map[string][]int64),
		stats:               make(map[string]models.TSStats),
		access:              &sync.Mutex{},

		history:    history,
		states:     states,
		buffersize: buffersize,
	}
}

// Counter takes a discretized TSState and counts the transition matrix
type Counter struct {
	// upper level profiler
	profiler api.TSProfiler

	// state
	currentState        map[string][]models.State
	stateChangeCounters map[string]map[string][]int64
	stats               map[string]models.TSStats
	access              *sync.Mutex

	// configs
	history    int
	states     int
	buffersize int
}

// Count takes a discretized Buffer represented as TSStates for each
// metric and increases the counter
func (counter *Counter) Count(tsstates []models.TSState) {
	for _, tsstate := range tsstates {
		// for each metric, add the given TSState
		counter.count(tsstate)
	}
}

// count takes a tsstate from a single metric, while Count takes an array
func (counter *Counter) count(tsstate models.TSState) {
	counter.access.Lock()
	defer counter.access.Unlock()

	// consider only the current metric
	metric := tsstate.Metric

	// handle default statistics
	stats := tsstate.Statistics
	globalStats := counter.stats[metric]
	if counter.stats[metric].Min > stats.Min || counter.stats[metric].Max < stats.Max {
		// min/max changed? update tx matrix dimension
		counter.stateChangeCounters[metric] = utils.ChangeDimension(counter.stateChangeCounters[metric], counter.stats[metric], stats, counter.states)

		if globalStats.Min == -1 || globalStats.Min > stats.Min {
			globalStats.Min = stats.Min
		}
		if globalStats.Max < stats.Max {
			globalStats.Max = stats.Max
		}
	}

	// update global stats from incoming stats
	oldAvg := globalStats.Avg
	globalStats.Avg = stat.Mean(
		[]float64{oldAvg, stats.Avg},
		[]float64{float64(globalStats.Count), float64(stats.Count)},
	)
	globalStats.Count += stats.Count
	globalStats.StddevSum += stats.StddevSum
	globalStats.Stddev = math.Sqrt(globalStats.StddevSum / float64(globalStats.Count))
	counter.stats[metric] = globalStats

	// handle state transitioning
	_, ok := counter.currentState[metric]
	if !ok {
		counter.currentState[metric] = make([]models.State, counter.history)
	}
	previousState := counter.currentState[metric]
	for len(previousState) > 0 {
		// first, find the previous state path
		previousStateIdent := ""
		for _, state := range previousState {
			if previousStateIdent != "" {
				previousStateIdent = previousStateIdent + "-"
			}
			previousStateIdent = previousStateIdent + fmt.Sprintf("%d", state.Value)
		}

		// increase (and create if not exists) counter for previousStateIdent
		_, ok := counter.stateChangeCounters[metric][previousStateIdent]
		if !ok {
			counter.stateChangeCounters[metric][previousStateIdent] = make([]int64, counter.states)
		}
		counter.stateChangeCounters[metric][previousStateIdent][tsstate.State.Value]++
		previousState = previousState[1:] // remove the handled previous state
	}

	// update new current state (remove oldest, append new state)
	if len(counter.currentState[metric]) > 0 {
		counter.currentState[metric] = counter.currentState[metric][1:] // remove first item
	}
	counter.currentState[metric] = append(counter.currentState[metric], tsstate.State) // add new item at the end

}

// GetTx returns the probability matrix for each metric
func (counter *Counter) GetTx() []models.TSProfileMetric {
	counter.access.Lock()
	defer counter.access.Unlock()
	var metrics []models.TSProfileMetric
	for metric, stateChangeCounter := range counter.stateChangeCounters {

		stats := counter.stats[metric]
		maxCount := float64(stats.Count) / float64(counter.buffersize) // WHY / BUFFERSIZE??
		txmatrix := utils.ComputeProbabilities(stateChangeCounter, maxCount)
		// fmt.Printf("counter %+v, probs: %+v\n", metricProfiler.counts.stateChangeCounter, txmatrix)
		metrics = append(metrics, models.TSProfileMetric{
			Name:     metric,
			TXMatrix: txmatrix,
			Stats:    stats,
		})
	}
	return metrics
}

func (counter *Counter) GetStats() map[string]models.TSStats {
	counter.access.Lock()
	defer counter.access.Unlock()
	return counter.stats
}

// Reset clears the counters
func (counter *Counter) Reset() {
	counter.access.Lock()
	defer counter.access.Unlock()
	//counter.currentState = make(map[string][]models.State)
	//counter.stateChangeCounter = make(map[string]map[string][]int64)
}
