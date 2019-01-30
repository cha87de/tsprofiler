package utils

import "math"

func IsOutlier(value float64, avg float64, stddev float64, filterStddev int) bool {
	if avg == 0 || filterStddev == -1 {
		return false
	}
	diff := math.Abs(value - avg)
	return diff >= float64(filterStddev)*stddev
}
