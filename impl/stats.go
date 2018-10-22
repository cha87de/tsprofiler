package impl

import "gonum.org/v1/gonum/stat"

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
