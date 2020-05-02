package predictor

import (
	"fmt"
	"math"
)

// Likeliness ...
func (predictor *Predictor) Likeliness(currentState map[string]string, steps int) (map[string][]int, error) {
	output := make(map[string][]int)

	txMatrices := predictor.getTxMatrices()

	// for each metric ...
	for _, txMatrix := range txMatrices {
		metric := txMatrix.Metric
		transitions := txMatrix.Transitions
		if _, exists := output[metric]; !exists {
			output[metric] = make([]int, predictor.profile.Settings.States)
		}

		// select current state in transitions
		//currentState := predictor.currentState[metric]
		txStep := transitions[currentState[metric]]

		//output[metric] = txStep.NextStateProbs

		if steps > 1 {
			// go to next step
			for nextState, nextStateProb := range txStep.NextStateProbs {
				if nextStateProb <= 0 {
					// ignore if unlikely
					continue
				}
				// go towards this next step
				nextStateHist := make(map[string]string)
				nextStateHist[metric] = fmt.Sprintf("%d", nextState)

				nextStepProbs, _ := predictor.Likeliness(nextStateHist, steps-1)

				for x := range output[metric] {
					nextStepProb := float64(nextStepProbs[metric][x]) / float64(100)
					thisStepProb := float64(nextStateProb) / float64(100)
					prob := nextStepProb * thisStepProb

					output[metric][x] += int(math.Round(prob * float64(100)))
				}
			}
		} else {
			// no more steps, return nextStateProbs
			output[metric] = txStep.NextStateProbs
		}
	}

	return output, nil
}
