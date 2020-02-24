package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cha87de/tsprofiler/models"
)

// ChangeDimension transforms the given sourceMatrix from stats oldStats to the new shape specified in newStats
func ChangeDimension(sourceMatrix map[string][]int64, oldStats models.TSStats, newStats models.TSStats, states int) map[string][]int64 {
	targetMatrix := make(map[string][]int64)

	oldMin := oldStats.Min
	oldMax := oldStats.Max
	oldStateStepSize := float64(oldMax-oldMin) / float64(states)

	newMin := newStats.Min
	newMax := newStats.Max

	//fmt.Printf("change dimensions from (%f,%f) to (%f,%f)\n", oldMin, oldMax, newMin, newMax)

	if newMin > oldMin {
		//fmt.Printf("Error: newMin larger than oldMin\n")
		newMin = oldMin // reset
	}
	if newMax < oldMax {
		//fmt.Printf("Error: newMax lower than oldMax\n")
		newMax = oldMax // reset
	}

	for key := range sourceMatrix {
		var newKey string
		for j := range sourceMatrix[key] {
			oldCounter := sourceMatrix[key][j]
			// were there any occurrences at all?
			if oldCounter <= 0 {
				continue
			}

			if newKey == "" {
				// lazy compute: state for i not yet calculated
				keyParts := strings.Split(key, "-")
				for _, keyPart := range keyParts {
					i, err := strconv.ParseInt(keyPart, 10, 32)
					if err != nil {
						i = 0
					}
					valueIpart := float64(i) * oldStateStepSize
					valueIpart += oldMin
					newStateIpart := ClosestDiscretize(valueIpart, states, newMin, newMax)
					if newStateIpart.Value < 0 || newStateIpart.Value >= int64(states) {
						fmt.Fprintf(os.Stderr, "no valid state found (iI). %.0f + %.0f * %s = %.0f (min %v, max %v, oldmin %v, oldmax %v)\n", oldMin, oldStateStepSize, key, valueIpart, newMin, newMax, oldMin, oldMax)
						// no state found
						newKey = ""
						break
					}
					if newKey != "" {
						newKey = newKey + "-"
					}
					newKey = newKey + fmt.Sprintf("%d", newStateIpart.Value)
				}
			}
			if newKey == "" {
				// if still empty, we have invalid states
				continue
			}
			valueJ := float64(j) * oldStateStepSize
			valueJ += oldMin
			newStateJ := ClosestDiscretize(valueJ, states, newMin, newMax)

			if newStateJ.Value < 0 || newStateJ.Value >= int64(states) {
				fmt.Fprintf(os.Stderr, "no valid state found (iJ) for value %v (min: %v, max %v, j: %v, stepsize: %v)\n", valueJ, newMin, newMax, j, oldStateStepSize)
				// no state found
				continue
			}
			//fmt.Printf("%+v,%+v\n", newStateI.value, newStateJ.value)
			_, ok := targetMatrix[newKey]
			if !ok {
				targetMatrix[newKey] = make([]int64, states)
			}
			targetMatrix[newKey][newStateJ.Value] += oldCounter
		}
	}
	return targetMatrix
}
