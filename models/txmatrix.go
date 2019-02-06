package models

import "github.com/cha87de/tsprofiler/utils"

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
	return utils.Round(ratio*1000) / 1000 // only 4 decimals please
}
