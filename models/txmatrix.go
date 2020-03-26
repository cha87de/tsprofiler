package models

import (
	"fmt"
	"math"
)

// TxMatrix describes for one metric a statistical profile
type TxMatrix struct {
	Metric      string            `json:"metric"`
	Transitions map[string]TXStep `json:"transitions"`
	Stats       TSStats           `json:"stats"`
}

// Diff compares two txMatrizes and returns the diff ratio between 0 (not equal) and 1 (fully equal)
func (txMatrix *TxMatrix) Diff(txMatrixRemote TxMatrix) float64 {
	counter := 0
	diffs := 0
	for state, txStep := range txMatrix.Transitions {
		remoteTxStep, ok := txMatrixRemote.Transitions[state]
		// counter = counter + 200 // maximal possible 2 * 100
		for i, nextStateProb := range txStep.NextStateProbs {
			counter = counter + nextStateProb
			if ok && len(remoteTxStep.NextStateProbs) > i {
				counter = counter + remoteTxStep.NextStateProbs[i]
				diff := nextStateProb - remoteTxStep.NextStateProbs[i]
				if diff < 0 {
					diff = diff * -1
				}
				if diff > counter {
					// max diff equals counter
					diff = counter
				}
				diffs = diffs + diff
			} else {
				// remote tx does not match. count as diff
				diffs = diffs + nextStateProb
			}
		}
	}
	ratio := float64(1) - float64(diffs)/float64(counter)
	return round(ratio*1000) / 1000 // only 4 decimals please
}

// Merge merges the given TxMatrix to the current one via average on the probabilities
func (txMatrix *TxMatrix) Merge(txMatrixRemote TxMatrix) {
	for state, txStep := range txMatrix.Transitions {
		remoteTxStep, ok := txMatrixRemote.Transitions[state]
		for i, nextStateProb := range txStep.NextStateProbs {
			if ok && len(remoteTxStep.NextStateProbs) > i {
				x := float64(nextStateProb)
				y := float64(remoteTxStep.NextStateProbs[i])
				z := int(round((x + y) / 2))
				// fmt.Printf("%f and %f = %d\n", x, y, z)
				txMatrix.Transitions[state].NextStateProbs[i] = z
			}
		}
	}
}

// Likeliness computes the likeliness for transitioning from the from state to the to state
func (txMatrix *TxMatrix) Likeliness(from []TSState, to TSState) float32 {

	fromIndex := fromString(from)
	for len(from) > 1 {
		fromIndex = fromString(from)
		if _, ok := txMatrix.Transitions[fromIndex]; ok {
			// found history
			break
		} else {
			// cut history
			from = from[1:]
		}
	}

	probs, ok := txMatrix.Transitions[fromIndex]
	if !ok {
		//fmt.Printf("from state %+v not found\n", from)
		return 0
	}
	if int(to.State.Value) > len(probs.NextStateProbs) {
		fmt.Printf("cannot compute likeliness: to state not existent")
		return 0
	}
	toProb := probs.NextStateProbs[to.State.Value]

	return float32(toProb) / 100
}

func round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}

func fromString(from []TSState) string {
	fromIndex := ""
	for _, s := range from {
		if fromIndex != "" {
			fromIndex = fromIndex + "-"
		}
		fromIndex = fromIndex + fmt.Sprintf("%d", s.State.Value)
	}
	return fromIndex
}

// TxLikeliness computes the likeliness that history happens on multivariate txMatrix
func TxLikeliness(txMatrices []TxMatrix, history [][]TSState, nextState []TSState) float32 {
	likelinessSum := float32(0)
	likelinessCount := 0

	for _, phaseTx := range txMatrices {

		fromStates := make([]TSState, 0)
		var toState TSState

		// find from
		for _, oldstateHistory := range history {
			for _, s := range oldstateHistory {
				if s.Metric == phaseTx.Metric {
					fromStates = append(fromStates, s)
					break
				}
			}
		}

		// find to
		for _, s := range nextState {
			if s.Metric == phaseTx.Metric {
				toState = s
				break
			}
		}

		likeliness := phaseTx.Likeliness(fromStates, toState)
		likelinessSum += likeliness
		likelinessCount++
	}

	return likelinessSum / float32(likelinessCount)
}
