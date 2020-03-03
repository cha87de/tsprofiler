package phase

import (
	"sync"

	"github.com/cha87de/tsprofiler/api"
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/profiler/counter"
)

const maxPhaseHistory = 5

func NewPhase(history int, states int, buffersize int, phaseLikeliness float32, phaseMin int64, profiler api.TSProfiler) Phase {
	phase := Phase{
		profiler: profiler,

		phaseCounters:        make([]counter.Counter, 1),
		phasePointer:         0,
		phaseTxCounter:       counter.NewCounter(1, 1, 1, profiler),
		phaseTSStatesHistory: make([][]models.TSState, 0),

		access: &sync.Mutex{},

		// config
		history:                  history,
		states:                   states,
		buffersize:               buffersize,
		phaseThresholdLikeliness: phaseLikeliness,
		phaseThresholdCounts:     phaseMin,
	}
	// create the first phase counter
	phase.phaseCounters[0] = counter.NewCounter(phase.history, phase.states, phase.buffersize, phase.profiler)

	return phase
}

type Phase struct {
	// upper level profiler
	profiler api.TSProfiler

	// state
	phaseCounters        []counter.Counter
	phasePointer         int
	phaseTxCounter       counter.Counter
	phaseTSStatesHistory [][]models.TSState

	access *sync.Mutex

	// configs
	history                  int
	states                   int
	buffersize               int
	phaseThresholdLikeliness float32
	phaseThresholdCounts     int64
}

func (phase *Phase) Count(tsstates []models.TSState) {
	phase.access.Lock()
	defer phase.access.Unlock()

	/*fmt.Printf("history: ")
	for _, n := range phase.phaseTSStatesHistory {
		for _, k := range n {
			fmt.Printf("%d", k.State.Value)
		}
		fmt.Printf(" ")
	}
	fmt.Printf("\n")*/

	likeliness := phase.phaseCounters[phase.phasePointer].Likeliness(tsstates)
	counts := phase.phaseCounters[phase.phasePointer].Totalcounts()
	//fmt.Printf("likeliness: %.2f, counts: %d\n", likeliness, counts)
	if likeliness < phase.phaseThresholdLikeliness && counts > phase.phaseThresholdCounts {
		// if likeliness is below threshold, look for better matching phase
		//fmt.Printf("likeliness: %.2f, counts: %d\n", likeliness, counts)

		// look for other phases
		newPhasePointer := -1
		for i, phaseCounter := range phase.phaseCounters {
			/*l := phase.Likeliness(tsstates)
			fmt.Printf("phase %d likeliness: %.2f\n", i, l)
			if l > 0 && likeliness < l {
				newPhasePointer = i
				likeliness = l
			}*/

			txMatrices := phaseCounter.GetTx()
			history := phase.phaseTSStatesHistory[:len(phase.phaseTSStatesHistory)-1]
			// nextState := tsstates
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
					nextState = tsstates
				}
				l := models.TxLikeliness(txMatrices, [][]models.TSState{historyStep}, nextState)
				//fmt.Printf("history step likeliness: %.2f\n", l)
				lSum += l
			}
			phaseLikeliness = lSum / float32(len(history))

			//fmt.Printf("phase %d likeliness: %.2f\n", i, phaseLikeliness)
			if likeliness < phaseLikeliness && phaseLikeliness > phase.phaseThresholdLikeliness {
				newPhasePointer = i
				likeliness = phaseLikeliness
			}

		}
		if newPhasePointer != -1 {
			// found a phase!
			//fmt.Printf("found matching phase %d\n", newPhasePointer)
			phase.phasePointer = newPhasePointer
		}

		// create a new phase
		if newPhasePointer == -1 {
			//fmt.Printf("create new phase\n")
			phase.phaseCounters = append(phase.phaseCounters, counter.NewCounter(phase.history, phase.states, phase.buffersize, phase.profiler))
			phase.phasePointer = len(phase.phaseCounters) - 1 // point to the newly added
		}
	}

	// increase counter on found phase
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
	if len(phase.phaseTSStatesHistory) > maxPhaseHistory {
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
