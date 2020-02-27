package task

import (
	"fmt"

	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/predictor"
)

// Simulate represents the simulation task of tspredictor
type Simulate struct {
	profile    models.TSProfile
	mode       predictor.PredictionMode
	simulation [][]models.TSState
	history    models.History
	//startStep  map[string]string
}

// NewSimulate creates and returns a new Simulate task
func NewSimulate(profile models.TSProfile, mode predictor.PredictionMode, history models.History) *Simulate {
	return &Simulate{
		profile:    profile,
		mode:       mode,
		simulation: make([][]models.TSState, 0),
		history:    history,
	}
}

// Run simulates given amount of steps
func (simulate *Simulate) Run(steps int) {
	predictor := predictor.NewPredictor(simulate.profile)
	predictor.SetMode(simulate.mode)
	// set last state
	lastState := simulate.history.HistoricStates[len(simulate.history.HistoricStates)-1]
	predictor.SetState(lastState)
	// set phase
	predictor.SetPhase(simulate.history.CurrentPhase)
	// set period path
	predictor.SetPeriodPath(simulate.history.PeriodPath, simulate.history.PeriodPathDepth)
	simulate.simulation = predictor.Simulate(steps)
}

// Print prints the simulation results to stdout
func (simulate *Simulate) Print() {
	if len(simulate.simulation) <= 0 {
		return
	}

	// print header
	for i, tsstate := range simulate.simulation[0] {
		if i > 0 {
			fmt.Printf(",")
		}
		fmt.Printf("%s", tsstate.Metric)
	}
	fmt.Printf("\n")

	// print rows
	for _, simstep := range simulate.simulation {
		for i, tsstate := range simstep {
			if i > 0 {
				fmt.Printf(",")
			}
			fmt.Printf("%d", tsstate.State.Value)
		}
		fmt.Printf("\n")
	}
}
