package period

import (
	"fmt"
	"sync"

	"github.com/cha87de/tsprofiler/api"
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/profiler/counter"
)

// NewPeriod initializes and returns a new Period
func NewPeriod(history int, states int, buffersize int, periodSize []int, profiler api.TSProfiler) Period {
	period := Period{
		profiler: profiler,

		// counters
		overallCounter: counter.NewCounter(history, states, buffersize, profiler),

		periodCounters:    make([]counter.Counter, len(periodSize)),
		periodSizeCounter: make([]int, len(periodSize)),

		phaseCounters:  make([]counter.Counter, 1),
		phasePointer:   0,
		phaseTxCounter: counter.NewCounter(1, 1, 1, profiler),

		access: &sync.Mutex{},

		// tx profiles as tree
		txTree:         models.NewPeriodTree(periodSize),
		txTreePosition: make([]int, len(periodSize)),

		// configs
		history:    history,
		states:     states,
		buffersize: buffersize,
		periodSize: periodSize,

		phaseThresholdLikeliness: 0.6,
		phaseThresholdCounts:     60,
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

	periodCounters    []counter.Counter
	periodSizeCounter []int

	phaseCounters  []counter.Counter
	phasePointer   int
	phaseTxCounter counter.Counter

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

	period.overallCounter.Count(tsstates)
	likeliness := period.phaseCounters[period.phasePointer].Likeliness(tsstates)
	counts := period.phaseCounters[period.phasePointer].Totalcounts()
	if likeliness < period.phaseThresholdLikeliness && counts > period.phaseThresholdCounts {
		// if likeliness is below threshold, look for better matching phase
		// fmt.Printf("likeliness: %.2f, counts: %d\n", likeliness, counts)

		// look for other phases
		newPhasePointer := -1
		for i, phase := range period.phaseCounters {
			l := phase.Likeliness(tsstates)
			if l > likeliness {
				newPhasePointer = i
				likeliness = l
			}
		}
		if newPhasePointer != -1 {
			// found a phase!
			// fmt.Printf("found matching phase %d\n", newPhasePointer)
			period.phasePointer = newPhasePointer
		}

		// create a new phase
		if newPhasePointer == -1 {
			period.phaseCounters = append(period.phaseCounters, counter.NewCounter(period.history, period.states, period.buffersize, period.profiler))
			period.phasePointer = len(period.phaseCounters) - 1 // point to the newly added
		}
	}
	period.phaseCounters[period.phasePointer].Count(tsstates)
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
	return models.Phases{
		Phases: txs,   // the list of detected phases
		Tx:     tx[0], // phase tx has only one metric by design
	}
}

// GetStats returns the first period's counter statistics
func (period *Period) GetStats() map[string]models.TSStats {
	// take period's overall counter
	return period.overallCounter.GetStats()
}
