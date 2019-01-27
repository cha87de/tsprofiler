package utils

import (
	"github.com/cha87de/tsprofiler/models"
)

func computeProbabilities(statematrix map[string][]int64, maxCount float64) map[string]models.TXStep {
	var output map[string]models.TXStep
	output = make(map[string]models.TXStep)
	for key, row := range statematrix {
		sum := Sum(row)
		var rowPerc []int
		for _, v := range row {
			var frac float64
			if sum == 0 {
				frac = 0.0
			} else {
				frac = float64(v) / float64(sum) * 100
			}
			fracInt := int(Round(frac))
			rowPerc = append(rowPerc, fracInt)
		}
		stepProb := float64(sum) / maxCount * 100
		output[key] = models.TXStep{
			NextStateProbs: rowPerc,
			StepProb:       int(Round(stepProb)),
		}

	}
	return output
}
