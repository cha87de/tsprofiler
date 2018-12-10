package impl

import (
	"math"
	"sort"

	"github.com/cha87de/tsprofiler/spec"
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
	if min == 0 && max == 0 {
		stateValue = 0
	}
	return state{
		value: stateValue,
	}
}

func computeProbabilities(statematrix map[string][]int64, maxCount float64) map[string]spec.TXStep {
	var output map[string]spec.TXStep
	output = make(map[string]spec.TXStep)
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
		stepProb := float64(sum) / maxCount * 100
		output[key] = spec.TXStep{
			NextStateProbs: rowPerc,
			StepProb:       int(round(stepProb)),
		}

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
