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
		//period.countPeriodTreeNode(tsstates)
	}
}

func (period *Period) countPeriodTreeNode(tsstates []models.TSState, level int) bool {
	//rootLevel := (level == 0)
	//leafLevel := (level == len(period.txTreePosition)-1)

	// now walk through the tree
	if level < len(period.txTreePosition)-1 {

		// always first count for current level.
		period.countPeriodTreeNodeLevel(tsstates, level)

		// go deeper into tree
		stepForward := period.countPeriodTreeNode(tsstates, level+1)

		if stepForward {
			// child level moved on
			period.txTreePosition[level]++
			//period.periodSizeCounter[level]++

			if period.txTreePosition[level] >= period.periodSize[level] {
				// yes! rotate and start from 0
				period.txTreePosition[level] = 0
				//period.periodSizeCounter[level] = 0
				// no! move on! clear level
				return true
			}

			// TODO: get counter from existing profile!!
			//treePos := period.txTreePosition[:level+1]
			//node := period.txTree.GetNode(treePos)
			//period.periodCounters[level] = createCounterFromTxMatrix(node.TxMatrix)
			period.periodCounters[level].ResetCounters()
			period.periodCounters[level].ResetStats()
		}
	} else { // level >= len(period.txTreePosition) ==> leaf node
		// we are on leaf level
		period.periodSizeCounter[level]++
		// can counter still be increased?
		if period.periodSizeCounter[level] >= period.periodSize[level] {
			// no! move on! clear level

			// TODO: get counter from existing profile!!
			//treePos := period.txTreePosition[:level+1]
			//node := period.txTree.GetNode(treePos)
			//period.periodCounters[level] = createCounterFromTxMatrix(node.TxMatrix)
			period.periodCounters[level].ResetCounters()
			period.periodCounters[level].ResetStats()

			period.periodSizeCounter[level] = 0
			return true
		}
	}
	return false
}

func (period *Period) countPeriodTreeNodeLevel(tsstates []models.TSState, level int) {
	period.periodCounters[level].Count(tsstates)

	// update tx
	tx := period.periodCounters[level].GetTx()
	treePos := period.txTreePosition[:level+1]
	node := period.txTree.GetNode(treePos)
	fmt.Printf("GetNode %+v (%d)\n", treePos, node.UUID)

	txMatrix := node.TxMatrix
	// if tx lengths unequal, overwrite (should never happen with proper input)
	if len(txMatrix) != len(tx) {
		txMatrix = tx
	} else {

		// TODO when rows 110 and 124 fixed, remove merging!

		// merge for each metric separately
		for m := range tx {
			txMatrix[m].Merge(tx[m])
			// merge stats
			txMatrix[m].Stats.Count++
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
	node.TxMatrix = txMatrix
}

/*
func (period *Period) countPeriodTreeNodeIterative(tsstates []models.TSState) {
	// for each tree level...
	for level := len(period.txTreePosition) - 1; level >= 0; level-- {

		//rootLevel := (level == 0)
		//leafLevel := (level == len(period.txTreePosition)-1)

		// get current node
		treePos := period.txTreePosition[:level+1]
		node := period.txTree.GetNode(treePos)

		// will counter increase exceed maxChilds?
		if period.periodSizeCounter[level]+1 >= node.MaxCounts {
			// yes! move on tree position on level to next child

			// will position increase exceed max. positions?
			if period.txTreePosition[level]+1 >= period.periodSize[level] {
				// yes! rotate and start from 0
				period.txTreePosition[level] = 0
			} else {
				// no! great, move on to next position on same level
				period.txTreePosition[level]++
			}
			// reset counters on level
			period.periodCounters[level].ResetCounters()
			period.periodCounters[level].ResetStats()
			period.periodSizeCounter[level] = 0
			// update current node
			treePos = period.txTreePosition[:level+1]
			node = period.txTree.GetNode(treePos)
		} else {
			// no problem, increase counter
			period.periodSizeCounter[level]++
		}

		period.periodCounters[level].Count(tsstates)

		fmt.Printf("level %d:\t treePos %d \t counter: %d \t node: %d (%+v)\n", level, period.txTreePosition[level], period.periodSizeCounter[level], node.UUID, period.txTreePosition)

		// update tx
		tx := period.periodCounters[level].GetTx()
		txMatrix := node.TxMatrix
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
		node.TxMatrix = txMatrix
	}
}

func (period *Period) countPeriodTreeNodeOld(tsstates []models.TSState, level int) bool {

	// handle tree
	moveOn := false
	nextLevel := level + 1
	if nextLevel < len(period.periodSize) {
		// go down into tree
		moveOn = period.countPeriodTreeNodeOld(tsstates, nextLevel)
	} else {
		// at leaf node level already
		//moveOn = (period.periodSizeCounter[level] >= period.periodSize[level]-1)
		treePos := period.txTreePosition[:level+1]
		node := period.txTree.Root.GetNode(treePos)
		moveOn = (period.periodSizeCounter[level] >= node.MaxCounts)
	}

	// check if running out of level
	moveOnUpperLevel := false
	if moveOn {
		//fmt.Printf("moveOn level %d\n", level)

		// time to move on ... is the current level full?
		period.txTreePosition[level] = period.txTreePosition[level] + 1
		treePos := period.txTreePosition[:level+1]
		node := period.txTree.Root.GetNode(treePos)
		if period.txTreePosition[level] >= node.MaxCounts {
			// reset position, start from 0  for current level
			period.txTreePosition[level] = 0
			moveOnUpperLevel = true
		}

		// reset for next node on tree level
		period.periodCounters[level].ResetCounters()
		period.periodCounters[level].ResetStats()
		period.periodSizeCounter[level] = 0
	}

	// count for current level
	period.periodCounters[level].Count(tsstates)
	period.periodSizeCounter[level]++
	// update tx
	tx := period.periodCounters[level].GetTx()
	treePos := period.txTreePosition[:level+1]
	node := period.txTree.Root.GetNode(treePos)
	fmt.Printf("treePos: %+v \t", treePos)
	var txMatrix []models.TxMatrix
	if level == 0 {
		// take rootTx
		txMatrix = period.txTree.Root.TxMatrix
		fmt.Printf("take rootTx\n")
	} else {
		// take children
		//treePos = treePos[:len(period.txTreePosition)-1]
		//node := period.txTree.Root.GetNode(treePos)
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

	return moveOnUpperLevel
}
*/

// GetTx returns for each period the counters' TSProfileMetric matrix
func (period *Period) GetTx() models.PeriodTree {
	return period.txTree
}

// GetCurrentPeriodPath returns the current tree positions
func (period *Period) GetCurrentPeriodPath() []int {
	return period.txTreePosition
}
