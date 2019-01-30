package utils

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/stat"
)

func Avg(data []float64) float64 {
	avg := stat.Mean(data, nil)
	return avg
}

func Stddev(data []float64) float64 {
	stddev := stat.StdDev(data, nil)
	return stddev
}

func Min(v []float64) float64 {
	sort.Float64s(v)
	return v[0]
}

func Max(v []float64) float64 {
	sort.Float64s(v)
	return v[len(v)-1]
}

func Sum(data []int64) int64 {
	var sum int64
	for _, v := range data {
		sum += v
	}
	return sum
}

func Round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}
