package task

import (
	"fmt"

	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/predictor"
)

// Likeliness represents the likeliness task of tspredictor
type Likeliness struct {
	profile    models.TSProfile
	mode       predictor.PredictionMode
	history    models.History
	likeliness map[string][]int
}

// NewLikeliness creates and returns a new Likeliness task
func NewLikeliness(profile models.TSProfile, mode predictor.PredictionMode, history models.History) *Likeliness {
	return &Likeliness{
		profile:    profile,
		mode:       mode,
		history:    history,
		likeliness: make(map[string][]int),
	}
}

// Run calculates likeliness for given next step
func (likeliness *Likeliness) Run(steps int, periodDepth int) error {
	var err error
	predictor := createPredictor(likeliness.profile, likeliness.mode, likeliness.history, periodDepth)
	currentState := likeliness.history.HistoricStates[len(likeliness.history.HistoricStates)-1]
	likeliness.likeliness, err = predictor.Likeliness(currentState, steps)
	if err != nil {
		return err
	}
	return nil
}

// Print prints the likeliness results to stdout
func (likeliness *Likeliness) Print() {
	if len(likeliness.likeliness) <= 0 {
		return
	}

	// print header
	fmt.Printf("state")
	for metric := range likeliness.likeliness {
		fmt.Printf(",")
		fmt.Printf("%s", metric)
	}
	fmt.Printf("\n")

	// print rows, each state one row
	for state := 0; state < likeliness.profile.Settings.States; state++ {
		fmt.Printf("%d", state)
		for _, l := range likeliness.likeliness {
			fmt.Printf(",")
			fmt.Printf("%d", l[state])
		}
		fmt.Printf("\n")
	}
}
