package utils

import (
	"fmt"

	"github.com/cha87de/tsprofiler/models"
)

// HistoryStateAsString returns the given state history as a string
func HistoryStateAsString(previousState []models.State) string {
	previousStateIdent := ""
	// first, find the previous state path
	for _, state := range previousState {
		if previousStateIdent != "" {
			previousStateIdent = previousStateIdent + "-"
		}
		previousStateIdent = previousStateIdent + fmt.Sprintf("%d", state.Value)
	}
	return previousStateIdent
}
