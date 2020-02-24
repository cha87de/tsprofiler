package period

import (
	"fmt"
	"sync"

	"github.com/cha87de/tsprofiler/api"
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/profiler/counter"
)

const maxPhaseHistory = 5

// NewPeriod initializes and returns a new Period
func NewPeriod(history int, states int, buffersize int, periodSize []int, phaseLikeliness float32, phaseMin int64, profiler api.TSProfiler) Period {
	period := Period{
		profiler: profiler,

		// counters
		overallCounter: counter.NewCounter(history, states, buffersize, profiler),
		lastStates:     make([]models.TSState, 0),

		periodCounters:    make([]counter.Counter, len(periodSize)),
		periodSizeCounter: make([]int, len(periodSize)),

		phaseCounters:        make([]counter.Counter, 1),
		phasePointer:         0,
		phaseTxCounter:       counter.NewCounter(1, 1, 1, profiler),
		phaseTSStatesHistory: make([][]models.TSState, 0),

		access: &sync.Mutex{},

		// tx profiles as tree
		txTree:         models.NewPeriodTree(periodSize),
		txTreePosition: make([]int, len(periodSize)),

		// configs
		history:                  history,
		states:                   states,
		buffersize:               buffersize,
		periodSize:               periodSize,
		phaseThresholdLikeliness: phaseLikeliness,
		phaseThresholdCounts:     phaseMin,
	}
	// create a counter for each entry in periodSize / for each level in PeriodTree
	for i := range periodSize {
		period.periodCounters[i] = counter.NewCounter(history, states, buffersize, profiler)
		period.periodSizeCounter[i] = 0
		period.txTreePosition[i] = 0
	}

	// create the first phase counter
	period.phaseCounters[0] = counter.NewCounter(period.history, period.states, period.buffersize, period.profiler)

	return period
}

// Period
type Period struct {
	// upper level profiler
	profiler api.TSProfiler

	// state
	overallCounter counter.Counter
	lastStates     []models.TSState

	periodCounters    []counter.Counter
	periodSizeCounter []int

	phaseCounters        []counter.Counter
	phasePointer         int
	phaseTxCounter       counter.Counter
	phaseTSStatesHistory [][]models.TSState

	txTree         models.PeriodTree
	txTreePosition []int

	access *sync.Mutex

	// configs
	history                  int
	states                   int
	buffersize               int
	periodSize               []int
	phaseThresholdLikeliness float32
	phaseThresholdCounts     int64
}

// Count takes a discretized Buffer represented as TSStates for each
// metric and increases the counter
func (period *Period) Count(tsstates []models.TSState) {
	period.access.Lock()
	defer period.access.Unlock()

	// global all time counting
	period.overallCounter.Count(tsstates)

	// Phase detection and counting
	period.countPhases(tsstates)

	// period tree counting
	period.countPeriodTree(tsstates)

	// update lastState
	period.lastStates = tsstates
}

func (period *Period) countPhases(tsstates []models.TSState) {

	fmt.Printf("history: ")
	for _, n := range period.phaseTSStatesHistory {
		for _, k := range n {
			fmt.Printf("%d", k.State.Value)
		}
		fmt.Printf(" ")
	}
	fmt.Printf("\n")

	likeliness := period.phaseCounters[period.phasePointer].Likeliness(tsstates)
	counts := period.phaseCounters[period.phasePointer].Totalcounts()
	fmt.Printf("likeliness: %.2f, counts: %d\n", likeliness, counts)
	if likeliness < period.phaseThresholdLikeliness && counts > period.phaseThresholdCounts {
		// if likeliness is below threshold, look for better matching phase
		//fmt.Printf("likeliness: %.2f, counts: %d\n", likeliness, counts)

		// look for other phases
		newPhasePointer := -1
		for i, phaseCounter := range period.phaseCounters {
			/*l := phase.Likeliness(tsstates)
			fmt.Printf("phase %d likeliness: %.2f\n", i, l)
			if l > 0 && likeliness < l {
				newPhasePointer = i
				likeliness = l
			}*/

			txMatrices := phaseCounter.GetTx()
			history := period.phaseTSStatesHistory[:len(period.phaseTSStatesHistory)-1]
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
				fmt.Printf("history step likeliness: %.2f\n", l)
				lSum += l
			}
			phaseLikeliness = lSum / float32(len(history))

			fmt.Printf("phase %d likeliness: %.2f\n", i, phaseLikeliness)
			if likeliness < phaseLikeliness && phaseLikeliness > period.phaseThresholdLikeliness {
				newPhasePointer = i
				likeliness = phaseLikeliness
			}

		}
		if newPhasePointer != -1 {
			// found a phase!
			fmt.Printf("found matching phase %d\n", newPhasePointer)
			period.phasePointer = newPhasePointer
		}

		// create a new phase
		if newPhasePointer == -1 {
			fmt.Printf("create new phase\n")
			period.phaseCounters = append(period.phaseCounters, counter.NewCounter(period.history, period.states, period.buffersize, period.profiler))
			period.phasePointer = len(period.phaseCounters) - 1 // point to the newly added
		}
	}

	// increase counter on found phase
	period.phaseCounters[period.phasePointer].Count(tsstates)

	// increase phase to phase counter
	phaseTsstates := make([]models.TSState, 1)
	phaseTsstates[0] = models.TSState{
		Metric: "phasetx",
		State: models.State{
			Value: int64(period.phasePointer),
		},
		Statistics: models.TSStats{
			Min:       0,
			Max:       float64(len(period.phaseCounters)),
			Stddev:    0,
			Avg:       0,
			Count:     1,
			StddevSum: 0,
		},
	}
	period.phaseTxCounter.Update(len(period.phaseCounters))
	period.phaseTxCounter.Count(phaseTsstates)

	// update history
	period.phaseTSStatesHistory = append(period.phaseTSStatesHistory, tsstates)
	if len(period.phaseTSStatesHistory) > maxPhaseHistory {
		// remove first (oldest) item
		period.phaseTSStatesHistory = period.phaseTSStatesHistory[1:]
	}
}

func (period *Period) countPeriodTree(tsstates []models.TSState) {

	// count for each period
	for i, size := range period.periodSize {
		counter := period.periodCounters[i]
		counter.Count(tsstates)
		period.periodSizeCounter[i]++

		if period.periodSizeCounter[i] >= size {
			tx := counter.GetTx()
			/*
				period full!

				- if no copy present, copy TSProfileMetric to this period txPerPeriod[i]
				- if copy is present, check how it differs from the current counter.GetTx()
				- if differs a lot, alert (if the copy is considered as stable)
				- if it differs a bit or copy is not stable, merge current tx with copy
			*/

			x := period.txTreePosition
			treePos := x[:len(period.txTreePosition)-i]
			node := period.txTree.GetNode(treePos)

			// alert or merge, depending on diff
			if len(node.TxMatrix) != len(tx) {
				node.TxMatrix = tx
			}
			// merge for each metric separately
			for m := range tx {
				localDiff := node.TxMatrix[m].Diff(tx[m])
				// 1.0 means equal, 0.0 means not equal
				fmt.Printf("localDiff is %.4f\n", localDiff)
				// if localDiff > float64(0.8) {
				node.TxMatrix[m].Merge(tx[m])
				// fmt.Printf("%+v", node.TxMatrix[m].Transitions)
				// } else {
				//	fmt.Printf("ALERT: localDiff is %.4f\n", localDiff)
				//}
			}

			// update tree position pointer
			period.txTreePosition[i] = period.txTreePosition[i] + 1
			if period.txTreePosition[i] >= size {
				// reset position, start from 0
				period.txTreePosition[i] = 0
			}

			counter.Reset()
			period.periodSizeCounter[i] = 0
		}
	}
}

// GetTx returns for each period the counters' TSProfileMetric matrix
func (period *Period) GetTx() models.PeriodTree {
	period.txTree.Root.TxMatrix = period.overallCounter.GetTx()
	return period.txTree
}

// GetPhasesTx returns
func (period *Period) GetPhasesTx() models.Phases {
	txs := make([][]models.TxMatrix, len(period.phaseCounters))
	for i, counter := range period.phaseCounters {
		phaseTx := counter.GetTx()
		txs[i] = phaseTx
	}
	tx := period.phaseTxCounter.GetTx()
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

// GetStats returns the first period's counter statistics
func (period *Period) GetStats() map[string]models.TSStats {
	// take period's overall counter
	return period.overallCounter.GetStats()
}

// GetState returns the last discretized state
func (period *Period) GetState() []models.TSState {
	return period.lastStates
}

// GetPhase returns the current phase pointer
func (period *Period) GetPhase() int {
	return period.phasePointer
}
