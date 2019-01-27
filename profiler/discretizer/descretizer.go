package discretizer

import (
	"fmt"

	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/utils"
)

// NewDiscretizer creates a new instance of Discretizer
func NewDiscretizer(states int) Discretizer {
	return Discretizer{
		states: states,
	}
}

// Discretizer computes from a TSBuffer a TSState
type Discretizer struct {
	// configs
	states int
}

// Discretize performs the computation of a discrete state from a TSBuffer
func (discretizer *Discretizer) Discretize(buffers []models.TSBuffer) []models.TSState {
	states := make([]models.TSState, len(buffers))
	// for each metric ...
	for i, buffer := range buffers {
		stats := discretizer.computeStats(buffer, currentAvg)
		state := utils.Discretize(stats.Avg, discretizer.states, stats.Min, stats.Max)
		if state.Value < 0 || state.Value >= int64(discretizer.states) {
			fmt.Printf("no valid state found (i) for value %v\n", stats.Avg)
			// no state found
			continue
		}
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
	stats.Max = buffer.Max
	stats.Min = buffer.Min
	stats.Stddev = utils.Stddev(buffer.RawData)
	stddevSum := float64(0)
	for _, v := range buffer.RawData {
		stddevSum += (v - currentAvg) * (v - stats.Avg)
	}
	stats.StddevSum = stddevSum
}
