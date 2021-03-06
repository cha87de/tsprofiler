package predictor

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/cha87de/tsprofiler/models"
	"github.com/jmcvetta/randutil"
)

func findMetricInTxMatrices(txmatrices []models.TxMatrix, metric string) (models.TxMatrix, error) {
	var txmatrix models.TxMatrix
	foundtx := false
	for _, tx := range txmatrices {
		if tx.Metric == metric {
			txmatrix = tx
			foundtx = true
		}
	}
	if !foundtx {
		err := fmt.Errorf("metric %s not found in TxMatrix array", metric)
		return txmatrix, err
	}
	return txmatrix, nil
}

func findStateHistoryInTxMatrix(txmatrix models.TxMatrix, stateHistory string) (models.TXStep, error) {
	var txstep models.TXStep
	stateHistoryArr := strings.Split(stateHistory, "-")
	found := false
	for len(stateHistoryArr) > 0 && !found {
		txstep, found = txmatrix.Transitions[strings.Join(stateHistoryArr, "-")]
		if !found {
			// cut stateHistory until we have a matching transitions
			stateHistoryArr = stateHistoryArr[1:]
		}
	}
	if !found {
		err := fmt.Errorf("cannot find state history %s in txmatrix", stateHistory)
		return txstep, err
	}
	return txstep, nil
}

func findStateByStateProbInTxmatrix(txmatrix models.TxMatrix) (models.TXStep, error) {
	// find state with highest probability
	state := ""
	stepProb := 0
	for s, txstep := range txmatrix.Transitions {
		if txstep.StepProb > stepProb {
			state = s
			stepProb = txstep.StepProb
		}
	}
	if state == "" {
		err := fmt.Errorf("failed to initialize state for metric %s", txmatrix.Metric)
		return models.TXStep{}, err
	}
	// fint txmatrix for new state
	step, err := findStateHistoryInTxMatrix(txmatrix, state)
	if err != nil {
		return models.TXStep{}, err
	}
	return step, nil
}

func computeNextState(nextStateProbs []int) (int, error) {
	choices := make([]randutil.Choice, len(nextStateProbs))
	for i, n := range nextStateProbs {
		choices[i] = randutil.Choice{
			Weight: n,
			Item:   i,
		}
	}

	result, err := randutil.WeightedChoice(choices)
	if err != nil {
		return 0, err
	}

	return result.Item.(int), nil
}

func computeValueFromState(state int, states int, min float64, max float64, stddev float64) int64 {
	stateSize := math.Round(float64(max-min) / float64(states))
	if stateSize <= 0 {
		//fmt.Printf("stateSize 0?!")
		return int64(0)
	}
	noise := float64(rand.Intn(int(stateSize))) * (stddev / max)
	value := min + float64(state)*stateSize + noise
	return int64(math.Round(value))
}
