package task

import (
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/predictor"
)

func createPredictor(profile models.TSProfile, mode predictor.PredictionMode, history models.History, periodDepth int) *predictor.Predictor {
	predictor := predictor.NewPredictor(profile)
	predictor.SetMode(mode)
	// set last state
	lastState := history.HistoricStates[len(history.HistoricStates)-1]
	predictor.SetState(lastState)
	// set phase
	predictor.SetPhase(history.CurrentPhase)
	// set period path
	predictor.SetPeriodPath(history.PeriodPath, periodDepth)

	return predictor
}
