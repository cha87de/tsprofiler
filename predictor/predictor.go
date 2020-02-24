package predictor

import (
	"fmt"
	"strings"

	"github.com/cha87de/tsprofiler/models"
)

// NewPredictor returns a new predictor for the given TSProfile
func NewPredictor(profile models.TSProfile) *Predictor {
	predictor := Predictor{
		profile:      profile,
		mode:         PredictionModeRootTx,
		currentPhase: 0,
	}
	predictor.initializeState()
	return &predictor
}

// Predictor offers prediction of the TSProfile the predictor is bound to
type Predictor struct {
	profile      models.TSProfile
	currentState map[string]string
	currentPhase int
	mode         PredictionMode
}

type nextState struct {
	state  int
	states int
	stats  models.TSStats
}

// NextState simulates next states for each metric using a random variable and the TSProfile's probabilities
func (predictor *Predictor) nextState() map[string]nextState {
	states := make(map[string]nextState)
	var txmatrices []models.TxMatrix

	// define which matrices to be used (default: root matrix)
	if predictor.mode == PredictionModeRootTx {
		txmatrices = predictor.profile.PeriodTree.Root.TxMatrix
	} else if predictor.mode == PredictionModePhases {
		predictor.nextPhase()
		txmatrices = predictor.profile.Phases.Phases[predictor.currentPhase]
	} else {
		fmt.Printf("warning: invalid prediction mode specified - falling back to root tx matrix")
		// fallback: root tx
		txmatrices = predictor.profile.PeriodTree.Root.TxMatrix
	}

	// for each metric
	for metric, stateHistory := range predictor.currentState {
		// find matrix for metric
		txmatrix, err := findMetricInTxMatrices(txmatrices, metric)
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		// find stateHistory in txmatrix
		txstep, err := findStateHistoryInTxMatrix(txmatrix, stateHistory)
		if err != nil {
			fmt.Printf("%s (phase %d, txmatrix %+v)\n", err, predictor.currentPhase, txmatrix)
			continue
		}

		// weighted random variable to define next state on txsteps
		next, err := computeNextState(txstep.NextStateProbs)
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		states[metric] = nextState{
			state:  next,
			states: predictor.profile.Settings.States,
			stats:  txmatrix.Stats,
		}
	}

	return states
}

func (predictor *Predictor) nextPhase() {
	currentPhase := predictor.currentPhase
	txmatrix := predictor.profile.Phases.Tx
	txstep, err := findStateHistoryInTxMatrix(txmatrix, fmt.Sprintf("%d", currentPhase))
	next, err := computeNextState(txstep.NextStateProbs)
	if err != nil {
		fmt.Printf("phase change error: %s\n", err)
		return
	}
	// fmt.Printf("phase now %d\n", next)
	predictor.currentPhase = next
	// when phase change, reset also history!
	if currentPhase != next {
		// fmt.Printf("phase change (%d -> %d), new init state\n", currentPhase, next)
		predictor.initializeState()
	}
}

// SetState defines the given currentState for the next simulation
func (predictor *Predictor) SetState(currentState map[string]string) {
	predictor.currentState = currentState
}

// SetPhase defines the given phase for the next simulation
func (predictor *Predictor) SetPhase(currentPhase int) {
	predictor.currentPhase = currentPhase
}

// SetMode defines the given PredictionMode for the next simulation
func (predictor *Predictor) SetMode(mode PredictionMode) {
	predictor.mode = mode
}

// Simulate computes `steps` states using randomness and TSProfile's probabilities
func (predictor *Predictor) Simulate(steps int) [][]models.TSState {
	simulation := make([][]models.TSState, steps)
	// fmt.Printf("start simulation with state %s\n", currentState)
	for i := 0; i < steps; i++ {
		next := predictor.nextState()
		j := 0
		simulation[i] = make([]models.TSState, len(next))
		nextStateHistory := make(map[string]string)
		for metric, state := range next {
			// compute value from state
			simValue := computeValueFromState(state.state, state.states, state.stats.Min, state.stats.Max, state.stats.Stddev)

			// pack value to array
			simulation[i][j] = models.TSState{
				Metric: metric,
				State: models.State{
					Value: simValue,
				},
			}
			j++

			// store next metric state for history
			nextStateHistory[metric] = fmt.Sprintf("%d", state.state)
		}
		predictor.appendState(nextStateHistory)
	}
	return simulation
}

// SimulateSteps computes `steps` states using randomness and TSProfile's probabilities
func (predictor *Predictor) SimulateSteps(steps int) [][]models.TSState {
	simulation := make([][]models.TSState, steps)
	// fmt.Printf("start simulation with state %s\n", currentState)
	for i := 0; i < steps; i++ {
		next := predictor.nextState()
		j := 0
		simulation[i] = make([]models.TSState, len(next))
		nextStateHistory := make(map[string]string)
		for metric, state := range next {
			// compute value from state
			simValue := int64(state.state)

			// pack value to array
			simulation[i][j] = models.TSState{
				Metric: metric,
				State: models.State{
					Value: simValue,
				},
			}
			j++

			// store next metric state for history
			nextStateHistory[metric] = fmt.Sprintf("%d", state.state)
		}
		predictor.appendState(nextStateHistory)
	}
	return simulation
}

func (predictor *Predictor) initializeState() {
	currentState := make(map[string]string)

	var txmatrices []models.TxMatrix
	if predictor.mode == PredictionModeRootTx {
		txmatrices = predictor.profile.PeriodTree.Root.TxMatrix
	} else if predictor.mode == PredictionModePhases {
		txmatrices = predictor.profile.Phases.Phases[predictor.currentPhase]
	} else {
		fmt.Printf("warning: invalid prediction mode specified - falling back to root tx matrix")
		// fallback: root tx
		txmatrices = predictor.profile.PeriodTree.Root.TxMatrix
	}
	for _, tx := range txmatrices {
		if _, exists := currentState[tx.Metric]; !exists {
			// find state with highest probability
			state := ""
			stepProb := 0
			for s, txstep := range tx.Transitions {
				if txstep.StepProb > stepProb {
					state = s
					stepProb = txstep.StepProb
				}
			}
			if state == "" {
				fmt.Printf("failed to initialize state for metric %s\n", tx.Metric)
				continue
			}
			currentState[tx.Metric] = state
		}
	}
	predictor.currentState = currentState
}

func (predictor *Predictor) appendState(state map[string]string) {
	for metric, state := range state {
		stateHistory, exists := predictor.currentState[metric]
		if !exists {
			// simply add current state
			predictor.currentState[metric] = state
			continue
		}

		stateHistoryArr := strings.Split(stateHistory, "-")

		// remove first (oldest) state (if max history reached)
		if len(stateHistoryArr) >= predictor.profile.Settings.History {
			stateHistoryArr = stateHistoryArr[1:]
		}

		// append last (newest) state
		stateHistoryArr = append(stateHistoryArr, state)

		// write back
		predictor.currentState[metric] = strings.Join(stateHistoryArr, "-")
	}
}
