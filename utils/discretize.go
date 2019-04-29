package utils

import (
	"github.com/cha87de/tsprofiler/models"
)

// SimpleDiscretize returns a state between min and max with maxstate steps of given value finding the smallest state
func SimpleDiscretize(value float64, maxstate int, min float64, max float64) models.State {
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

// ClosestDiscretize returns a state between min and max with maxstate steps of given value, finding the closest state
func ClosestDiscretize(value float64, maxstate int, min float64, max float64) models.State {
	stateStepSize := float64(max-min) / float64(maxstate)
	for i := 0; i < maxstate; i++ {
		lowerbound := float64(i)*stateStepSize - 0.5*stateStepSize
		upperbound := float64(i)*stateStepSize + 0.5*stateStepSize
		if value >= lowerbound && value < upperbound {
			return models.State{
				Value: int64(i),
			}
		}
	}
	maxupperbound := float64(maxstate-1)*stateStepSize + 0.5*stateStepSize
	if value >= maxupperbound {
		// exceeding the bound towards top
		return models.State{
			Value: int64(maxstate - 1),
		}
	}
	return models.State{
		Value: int64(0),
	}
}
