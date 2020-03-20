package period

import (
	"fmt"
	"math"
	"sync"

	"github.com/cha87de/tsprofiler/api"
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/profiler/counter"
	"gonum.org/v1/gonum/stat"
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
	if len(period.periodSize) > 0 {
		// only count period when configured
		//fmt.Printf("txTreePos: %+v\n", period.txTreePosition)
		period.countPeriodTreeNode(tsstates, 0)
	}
}

func (period *Period) countPeriodTreeNode(tsstates []models.TSState, level int) bool {

	// handle tree
	moveOn := false
	nextLevel := level + 1
	if nextLevel < len(period.periodSize) {
		// go down into tree
		moveOn = period.countPeriodTreeNode(tsstates, nextLevel)
	} else {
		// at leaf node level already
		// caution: periodSize starts at 1, counter starts at 0!!
		moveOn = (period.periodSizeCounter[level] >= period.periodSize[level]-1)
	}

	// count for current level
	period.periodCounters[level].Count(tsstates)
	period.periodSizeCounter[level]++

	// update tx
	tx := period.periodCounters[level].GetTx()
	treePos := period.txTreePosition[:len(period.txTreePosition)-level]
	fmt.Printf("treePos: %+v ", treePos)
	var txMatrix []models.TxMatrix
	if level == len(period.periodSize)-1 {
		// take rootTx
		txMatrix = period.txTree.Root.TxMatrix
		fmt.Printf("take rootTx\n")
	} else {
		// take children
		//treePos = treePos[:len(period.txTreePosition)-1]
		node := period.txTree.Root.GetNode(treePos)
		txMatrix = node.TxMatrix
		fmt.Printf("received node %d\n", node.UUID)
	}

	// if tx lengths unequal, overwrite
	if len(txMatrix) != len(tx) {
		txMatrix = tx
	} else {
		// merge for each metric separately
		for m := range tx {
			txMatrix[m].Merge(tx[m])
			// merge stats
			txMatrix[m].Stats.Count += tx[m].Stats.Count
			if txMatrix[m].Stats.Min > tx[m].Stats.Min {
				txMatrix[m].Stats.Min = tx[m].Stats.Min
			}
			if txMatrix[m].Stats.Max < tx[m].Stats.Max {
				txMatrix[m].Stats.Max = tx[m].Stats.Max
			}
			// avg
			mergedAvg := stat.Mean(
				[]float64{txMatrix[m].Stats.Avg, tx[m].Stats.Avg},
				[]float64{float64(txMatrix[m].Stats.Count), float64(tx[m].Stats.Count)},
			)
			txMatrix[m].Stats.Avg = mergedAvg
			// stddev
			txMatrix[m].Stats.StddevSum += tx[m].Stats.StddevSum
			txMatrix[m].Stats.Stddev = math.Sqrt(txMatrix[m].Stats.StddevSum / float64(txMatrix[m].Stats.Count))
		}
	}
	if level == len(period.periodSize)-1 {
		period.txTree.Root.TxMatrix = txMatrix
	} else {
		node := period.txTree.Root.GetNode(treePos)
		node.TxMatrix = txMatrix
	}

	// check if running out of level
	moveOnUpperLevel := false
	if moveOn {
		//fmt.Printf("moveOn level %d\n", level)

		// time to move on ... is the current level full?
		period.txTreePosition[level] = period.txTreePosition[level] + 1
		if period.txTreePosition[level] >= period.periodSize[level] {
			// reset position, start from 0  for current level
			period.txTreePosition[level] = 0
			moveOnUpperLevel = true
		}

		// reset for next node on tree level
		period.periodCounters[level].ResetCounters()
		period.periodCounters[level].ResetStats()
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
