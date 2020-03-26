package phase

import (
	"fmt"
	"math"
	"sync"

	"github.com/cha87de/tsprofiler/api"
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/profiler/counter"
)

// NewPhase instantiates and returns a new Phase with the provided parameters
func NewPhase(history int, states int, buffersize int, phaseLikeliness float32, phaseHistory int64, profiler api.TSProfiler) Phase {
	phase := Phase{
		profiler: profiler,

		phaseCounters:                  make([]counter.Counter, 1),
		phasePointer:                   0,
		phaseTxCounter:                 counter.NewCounter(1, 1, 1, profiler),
		phaseTSStatesHistory:           make([][]models.TSState, 0),
		phaseTSStatesHistoryLikeliness: make([]float32, 0),

		access: &sync.Mutex{},

		// config
		history:                  history,
		states:                   states,
		buffersize:               buffersize,
		phaseThresholdLikeliness: phaseLikeliness,
		phaseThresholdHistory:    phaseHistory,
	}
	// create the first phase counter
	phase.phaseCounters[0] = counter.NewCounter(phase.history, phase.states, phase.buffersize, phase.profiler)

	return phase
}

// Phase handles the phase detection and state counting of the profiler
type Phase struct {
	// upper level profiler
	profiler api.TSProfiler

	// state
	phaseCounters                  []counter.Counter
	phasePointer                   int
	phaseTxCounter                 counter.Counter
	phaseTSStatesHistory           [][]models.TSState
	phaseTSStatesHistoryLikeliness []float32

	access *sync.Mutex

	// configs
	history                  int
	states                   int
	buffersize               int
	phaseThresholdLikeliness float32
	phaseThresholdHistory    int64
}

// Count takes a discretized Buffer represented as TSStates for each metric,
// adjusts the current phase and increases its counter
func (phase *Phase) Count(tsstates []models.TSState) {
	phase.access.Lock()
	defer phase.access.Unlock()

	// update likeliness history
	currentLikeliness := phase.phaseCounters[phase.phasePointer].Likeliness(tsstates)
	if math.IsNaN(float64(currentLikeliness)) {
		currentLikeliness = 1
	}
	phase.phaseTSStatesHistoryLikeliness = append(phase.phaseTSStatesHistoryLikeliness, currentLikeliness)
	if int64(len(phase.phaseTSStatesHistoryLikeliness)) > phase.phaseThresholdHistory {
		// remove first (oldest) item
		phase.phaseTSStatesHistoryLikeliness = phase.phaseTSStatesHistoryLikeliness[1:]
	}

	// calculate historyLikeliness
	historyLikelinessSum := float32(0)
	//countSum := 0
	for _, likeliness := range phase.phaseTSStatesHistoryLikeliness {
		//historyLikelinessSum += likeliness * float32(i+1)
		historyLikelinessSum += likeliness
		//countSum += (i + 1)
	}
	//historyLikeliness := historyLikelinessSum / float32(countSum)
	historyLikeliness := historyLikelinessSum / float32(len(phase.phaseTSStatesHistoryLikeliness))
	//historyLikeliness := currentLikeliness

	if historyLikeliness < phase.phaseThresholdLikeliness {
		// if likeliness is below threshold, look for better matching phase!

		//fmt.Printf("likeliness: %.2f, counts: %d\n", likeliness, counts)
		//fmt.Printf("start lookup for other phase\n")

		// loop through phases to search better matching one
		newPhasePointer := -1
		for i, phaseCounter := range phase.phaseCounters {
			txMatrices := phaseCounter.GetTx()
			history := phase.phaseTSStatesHistory[:len(phase.phaseTSStatesHistory)-1]

			// caution: cannot provide history directly to TxLikeliness:
			// history with x,y,z elements would be the state "x-y-z"!

			var phaseLikeliness float32
			lSum := float32(0)
			for i, historyStep := range history {
				var nextState []models.TSState
				if (i + 1) < len(history) {
					// as long as history has "next" item, take from history
					nextState = history[i+1]
				} else {
					// end of historic states, take currently incoming tsstate
					nextState = tsstates
				}
				l := models.TxLikeliness(txMatrices, [][]models.TSState{historyStep}, nextState)
				lSum += l
			}
			phaseLikeliness = lSum / float32(len(history))

			//fmt.Printf("phase %d likeliness: %.2f\n", i, phaseLikeliness)
			if historyLikeliness < phaseLikeliness && phaseLikeliness > phase.phaseThresholdLikeliness {
				newPhasePointer = i
				historyLikeliness = phaseLikeliness
			}
		}
		if newPhasePointer != -1 {
			// found an existing, matching phase!
			fmt.Printf("found better matching phase %d (%.3f)\n", newPhasePointer, historyLikeliness)
			phase.phasePointer = newPhasePointer
		} else {
			// create a new phase
			phaseid := len(phase.phaseCounters) - 1
			fmt.Printf("create new phase %d\n", phaseid)
			phase.phaseCounters = append(phase.phaseCounters, counter.NewCounter(phase.history, phase.states, phase.buffersize, phase.profiler))
			phase.phasePointer = phaseid // point to the newly added
		}
	}

	// increase counter on current phase
	phase.phaseCounters[phase.phasePointer].Count(tsstates)

	// increase phase to phase counter
	phaseTsstates := make([]models.TSState, 1)
	phaseTsstates[0] = models.TSState{
		Metric: "phasetx",
		State: models.State{
			Value: int64(phase.phasePointer),
		},
		Statistics: models.TSStats{
			Min:       0,
			Max:       float64(len(phase.phaseCounters)),
			Stddev:    0,
			Avg:       0,
			Count:     1,
			StddevSum: 0,
		},
	}
	phase.phaseTxCounter.Update(len(phase.phaseCounters))
	phase.phaseTxCounter.Count(phaseTsstates)

	// update history
	phase.phaseTSStatesHistory = append(phase.phaseTSStatesHistory, tsstates)
	if int64(len(phase.phaseTSStatesHistory)) > phase.phaseThresholdHistory {
		// remove first (oldest) item
		phase.phaseTSStatesHistory = phase.phaseTSStatesHistory[1:]
	}
}

// GetPhasesTx returns
func (phase *Phase) GetPhasesTx() models.Phases {
	txs := make([][]models.TxMatrix, len(phase.phaseCounters))
	for i, counter := range phase.phaseCounters {
		phaseTx := counter.GetTx()
		txs[i] = phaseTx
	}
	tx := phase.phaseTxCounter.GetTx()
	var txMetric models.TxMatrix
	if len(tx) > 0 {
		txMetric = tx[0]
		/*} else {
		fmt.Printf("tx metric 0 ?! wtf")*/
	}
	return models.Phases{
		Phases: txs,      // the list of detected phases
		Tx:     txMetric, // phase tx has only one metric by design
	}
}

// GetPhase returns the current phase pointer
func (phase *Phase) GetPhase() int {
	return phase.phasePointer
}
