package counter

import (
	"fmt"

	"github.com/cha87de/tsprofiler/models"
)

// NewCounter initializes and returns a new Counter
func NewCounter(history int) Counter {
	return Counter{
		history: history,
	}
}

// Counter takes a discretized TSState and counts the transition matrix
type Counter struct {
	// state
	currentState       map[string][]models.State
	stateChangeCounter map[string][]int64

	// configs
	history int
}

// Count takes a discretized Buffer represented as TSStates for each
// metric and increases the counter
func (counter *Counter) Count(tsstates []models.TSState) {
	// for each metric ...
	for _, tsstate := range tsstates {
		metric := tsstate.Metric
		previousState := counter.currentState[metric]
		for len(previousState) > 0 {
			previousStateIdent := ""
			for _, state := range previousState {
				if previousStateIdent != "" {
					previousStateIdent = previousStateIdent + "-"
				}
				previousStateIdent = previousStateIdent + fmt.Sprintf("%d", state.Value)
			}
			_, ok := profilerMetric.counts.stateChangeCounter[previousStateIdent]
			if !ok {
				profilerMetric.counts.stateChangeCounter[previousStateIdent] = make([]int64, profilerMetric.counts.maxstates)
			}
			profilerMetric.counts.stateChangeCounter[previousStateIdent][newState.value]++
			previousState = previousState[1:]
		}

		if len(profilerMetric.counts.currentState) > 0 {
			profilerMetric.counts.currentState = profilerMetric.counts.currentState[1:] // remove first item
		}
		profilerMetric.counts.currentState = append(profilerMetric.counts.currentState, newState) // add new item at the end

	}
}
