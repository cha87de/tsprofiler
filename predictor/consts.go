package predictor

// PredictionMode defines the mode used for the predictors simulation
type PredictionMode int

const (
	// PredictionModeRootTx defines the mode "RootTx", which uses the TSProfile's root transition matrix
	PredictionModeRootTx PredictionMode = 0

	// PredictionModePhases defines the mode "Phases", which uses the TSProfile's phases
	PredictionModePhases PredictionMode = 1

	// PredictionModePeriods defines the mode "RootTx", which uses the TSProfile's periods
	PredictionModePeriods PredictionMode = 2
)
