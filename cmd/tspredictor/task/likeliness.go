package task

import (
	"fmt"

	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/predictor"
)

// Likeliness represents the likeliness task of tspredictor
type Likeliness struct {
	profile models.TSProfile
	mode    predictor.PredictionMode
	history models.History
}

// NewLikeliness creates and returns a new Likeliness task
func NewLikeliness(profile models.TSProfile, mode predictor.PredictionMode, history models.History) *Likeliness {
	return &Likeliness{
		profile: profile,
		mode:    mode,
		history: history,
	}
}

// Run calculates likeliness for given next step
func (likeliness *Likeliness) Run(periodDepth int) error {

	return fmt.Errorf("Task Likeliness not implement")

	//predictor := createPredictor(likeliness.profile, likeliness.mode, likeliness.history, periodDepth)

	//whatever, _ := predictor.SimulateSteps(1)

	//likeliness.history.NextState
	//fmt.Printf("%+v", whatever)

	//return nil
}
