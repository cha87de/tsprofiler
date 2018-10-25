package impl

import (
	"math"

	"gonum.org/v1/gonum/stat"
)

type state struct {
	value int64
}

const maxstates = 4

func aggregate(data []float64) float64 {
	avg := stat.Mean(data, nil)
	return avg
}

func sum(data []int64) int64 {
	var sum int64
	for _, v := range data {
		sum += v
	}
	return sum
}

func discretize(value float64) state {
	if value < 25 {
		return state{value: 0}
	} else if value < 50 {
		return state{value: 1}
	} else if value < 75 {
		return state{value: 2}
	} else if value <= 100 {
		return state{value: 3}
	}
	return state{value: 0}
}

func computeProbabilities(statematrix [][]int64) [][]int {
	var output [][]int
	for _, row := range statematrix {
		sum := sum(row)
		var rowPerc []int
		for _, v := range row {
			var frac float64
			if sum == 0 {
				frac = 0.0
			} else {
				frac = float64(v) / float64(sum) * 100
			}
			fracInt := int(math.Round(frac))
			rowPerc = append(rowPerc, fracInt)
		}
		output = append(output, rowPerc)
	}
	return output
}
