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
		overallCounter:    counter.NewCounter(history, states, buffersize, profiler),
		counters:          make([]counter.Counter, len(periodSize)),
		periodSizeCounter: make([]int, len(periodSize)),
		access:            &sync.Mutex{},

		// tx profiles as tree
		txTree:         models.NewPeriodTree(periodSize),
		txTreePosition: make([]int, len(periodSize)),

		periodSize: periodSize,
	}
	// create a counter for each entry in periodSize / for each level in PeriodTree
	for i := range periodSize {
		period.counters[i] = counter.NewCounter(history, states, buffersize, profiler)
		period.periodSizeCounter[i] = 0
		period.txTreePosition[i] = 0
	}

	return period
}

// Period
type Period struct {
	// upper level profiler
	profiler api.TSProfiler

	// state
	overallCounter    counter.Counter
	counters          []counter.Counter
	periodSizeCounter []int

	txTree         models.PeriodTree
	txTreePosition []int

	access *sync.Mutex

	// configs
	periodSize []int
}

// Count takes a discretized Buffer represented as TSStates for each
// metric and increases the counter
func (period *Period) Count(tsstates []models.TSState) {
	period.access.Lock()
	defer period.access.Unlock()

	period.overallCounter.Count(tsstates)

	// count for each period
	for i, size := range period.periodSize {
		counter := period.counters[i]
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
			for m := range tx {
				localDiff := node.TxMatrix[m].Diff(tx[m])
				// 1.0 means equal, 0.0 means not equal
				fmt.Printf("localDiff is %.4f\n", localDiff)
				// if localDiff > float64(0.8) {
				node.TxMatrix[m].Merge(tx[m])
				// fmt.Printf("%+v", node.TxMatrix[m].Transitions)
				// TODO MERGE INSTEAD OF overwrite
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

// GetStats returns the first period's counter statistics
func (period *Period) GetStats() map[string]models.TSStats {
	// take period's overall counter
	return period.overallCounter.GetStats()
}
