package discretizer

import (
	"fmt"
	"os"

	"github.com/cha87de/tsprofiler/api"
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/utils"
)

// NewDiscretizer creates a new instance of Discretizer
func NewDiscretizer(states int, fixedBound bool, profiler api.TSProfiler) Discretizer {
	return Discretizer{
		profiler:   profiler,
		states:     states,
		fixedBound: fixedBound,
	}
}

// Discretizer computes from a TSBuffer a TSState
type Discretizer struct {
	// upper level profiler
	profiler api.TSProfiler

	// configs
	states     int
	fixedBound bool

	// no state or cache required
}

// Discretize performs the computation of a discrete state from a TSBuffer
func (discretizer *Discretizer) Discretize(buffers []models.TSBuffer) []models.TSState {
	states := make([]models.TSState, len(buffers))
	currentStates := discretizer.profiler.GetCurrentState()
	// for each metric ...
	for i, buffer := range buffers {
		// find matching currentState
		currentState, currentStateFound := currentStates[buffer.Metric]
		var currentAvg float64
		if currentStateFound {
			currentAvg = currentState.Avg
		}

		// compute basic statistics
		stats := discretizer.computeStats(buffer, currentAvg)

		// compute state from basic statistics
		state := utils.Discretize(stats.Avg, discretizer.states, stats.Min, stats.Max)
		if state.Value < 0 || state.Value >= int64(discretizer.states) {
			fmt.Fprintf(os.Stderr, "no valid state found (i) for value %v\n", stats.Avg)
			// no state found
			continue
		}

		// bundle TSState
		states[i] = models.TSState{
			Metric:     buffer.Metric,
			State:      state,
			Statistics: stats,
		}
	}
	return states
}

func (discretizer *Discretizer) computeStats(buffer models.TSBuffer, currentAvg float64) models.TSStats {
	stats := models.TSStats{}
	stats.Avg = utils.Avg(buffer.RawData)
	stats.Count = int64(len(buffer.RawData))
	stats.Min = buffer.Min
	stats.Max = buffer.Max
	if discretizer.fixedBound {
		stats.Min = buffer.FixedMin
		stats.Max = buffer.FixedMax
	}
	stats.Stddev = utils.Stddev(buffer.RawData)
	stddevSum := float64(0)
	for _, v := range buffer.RawData {
		stddevSum += (v - currentAvg) * (v - stats.Avg)
	}
	stats.StddevSum = stddevSum
	return stats
}
