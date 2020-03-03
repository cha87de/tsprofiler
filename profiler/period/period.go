package period

import (
	"sync"

	"github.com/cha87de/tsprofiler/api"
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/profiler/counter"
)

// NewPeriod initializes and returns a new Period
func NewPeriod(history int, states int, buffersize int, periodSize []int, profiler api.TSProfiler) Period {
	period := Period{
		profiler: profiler,

		periodCounters:    make([]counter.Counter, len(periodSize)),
		periodSizeCounter: make([]int, len(periodSize)),

		access: &sync.Mutex{},

		// tx profiles as tree
		txTree:         models.NewPeriodTree(periodSize),
		txTreePosition: make([]int, len(periodSize)),

		// configs
		history:    history,
		states:     states,
		buffersize: buffersize,
		periodSize: periodSize,
	}
	// create a counter for each entry in periodSize / for each level in PeriodTree
	for i := range periodSize {
		period.periodCounters[i] = counter.NewCounter(history, states, buffersize, profiler)
		period.periodSizeCounter[i] = 0
		period.txTreePosition[i] = 0
	}

	return period
}

// Period holds counters etc to compute probabilities for given period size
type Period struct {
	// upper level profiler
	profiler api.TSProfiler

	// state
	periodCounters    []counter.Counter
	periodSizeCounter []int

	txTree         models.PeriodTree
	txTreePosition []int

	access *sync.Mutex

	// configs
	history    int
	states     int
	buffersize int
	periodSize []int
}

// Count takes a discretized Buffer represented as TSStates for each
// metric and increases the counter
func (period *Period) Count(tsstates []models.TSState) {
	period.access.Lock()
	defer period.access.Unlock()

	// period tree counting
	period.countPeriodTree(tsstates)
}

func (period *Period) countPeriodTree(tsstates []models.TSState) {
	period.countPeriodTreeNode(tsstates, 0)
}

func (period *Period) countPeriodTreeNode(tsstates []models.TSState, level int) bool {

	// count for current level
	period.periodCounters[level].Count(tsstates)
	period.periodSizeCounter[level]++

	// handle tree
	moveOn := false
	nextLevel := level + 1
	if nextLevel < len(period.periodSize) {
		// go down into tree
		nextLevel := level + 1
		moveOn = period.countPeriodTreeNode(tsstates, nextLevel)
	} else {
		// at leaf node level already
		moveOn = (period.periodSizeCounter[level] >= period.periodSize[level])
	}

	// always update tx
	tx := period.periodCounters[level].GetTx()
	treePos := period.txTreePosition[:len(period.txTreePosition)-level]
	node := period.txTree.GetNode(treePos)
	// if tx lengths unequal, overwrite
	if len(node.TxMatrix) != len(tx) {
		node.TxMatrix = tx
	} else {
		// merge for each metric separately
		for m := range tx {
			node.TxMatrix[m].Merge(tx[m])
			// TODO merge stats
		}
	}

	// check if running out of level
	moveOnUpperLevel := false
	if moveOn {
		// time to move on ... is the current level full?
		period.txTreePosition[level] = period.txTreePosition[level] + 1
		if period.txTreePosition[level] >= period.periodSize[level] {
			// reset position, start from 0  for current level
			period.txTreePosition[level] = 0
			moveOnUpperLevel = true
		}

		// reset for next node on tree level
		period.periodCounters[level].Reset()
		period.periodSizeCounter[level] = 0
	}

	return moveOnUpperLevel
}

// GetTx returns for each period the counters' TSProfileMetric matrix
func (period *Period) GetTx() models.PeriodTree {
	return period.txTree
}

// GetCurrentPeriodPath returns the current tree positions
func (period *Period) GetCurrentPeriodPath() []int {
	return period.txTreePosition
}
