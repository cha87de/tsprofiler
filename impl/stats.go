package impl

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/stat"
)

type state struct {
	value int64
}

func avg(data []float64) float64 {
	avg := stat.Mean(data, nil)
	return avg
}

func stddev(data []float64) float64 {
	stddev := stat.StdDev(data, nil)
	return stddev
}

func min(v []float64) float64 {
	sort.Float64s(v)
	return v[0]
}

func max(v []float64) float64 {
	sort.Float64s(v)
	return v[len(v)-1]
}

func sum(data []int64) int64 {
	var sum int64
	for _, v := range data {
		sum += v
	}
	return sum
}

func discretize(value float64, maxstate int, min float64, max float64) state {
	stateStepSize := float64(max-min) / float64(maxstate)
	stateStepValue := min
	stateValue := int64(-1)
	for stateStepValue < max {
		if value < stateStepValue {
			return state{
				value: stateValue,
			}
		}
		stateValue++
		stateStepValue += stateStepSize
	}
	return state{
		value: stateValue,
	}
}

func computeProbabilities(statematrix map[string][]int64) map[string][]int {
	var output map[string][]int
	output = make(map[string][]int)
	for key, row := range statematrix {
		sum := sum(row)
		var rowPerc []int
		for _, v := range row {
			var frac float64
			if sum == 0 {
				frac = 0.0
			} else {
				frac = float64(v) / float64(sum) * 100
			}
			fracInt := int(round(frac))
			rowPerc = append(rowPerc, fracInt)
		}
		//output = append(output, rowPerc)
		output[key] = rowPerc

	}
	return output
}

func round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}
