package utils

import "github.com/cha87de/tsprofiler/models"

// Discretize returns a state between min and max with maxstate steps of given value
func Discretize(value float64, maxstate int, min float64, max float64) models.State {
	stateStepSize := float64(max-min) / float64(maxstate)
	stateStepValue := min
	stateValue := int64(-1)
	for stateStepValue < max {
		if value < stateStepValue {
			return models.State{
				Value: stateValue,
			}
		}
		stateValue++
		stateStepValue += stateStepSize
	}
	if min == 0 && max == 0 {
		stateValue = 0
	}
	return models.State{
		Value: stateValue,
	}
}
