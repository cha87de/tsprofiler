package predictor

import (
	"fmt"
	"strings"

	"github.com/cha87de/tsprofiler/models"
)

// NewPredictor returns a new predictor for the given TSProfile
func NewPredictor(profile models.TSProfile) *Predictor {
	predictor := Predictor{
		profile:           profile,
		currentPhase:      0,
		periodSizeCounter: make([]int, len(profile.Settings.PeriodSize)),
	}
	predictor.initializeState()
	return &predictor
}

// Predictor offers prediction of the TSProfile the predictor is bound to
type Predictor struct {
	profile models.TSProfile

	currentState map[string]string
	currentPhase int

	periodPath        []int
	periodPathDepth   int
	periodSizeCounter []int

	mode PredictionMode
}

type nextState struct {
	state  int
	states int
	stats  models.TSStats
}

// NextState simulates next states for each metric using a random variable and the TSProfile's probabilities
func (predictor *Predictor) nextState() (map[string]nextState, error) {
	states := make(map[string]nextState)
	var txmatrices []models.TxMatrix

	// define which matrices to be used (default: root matrix)
	if predictor.mode == PredictionModeRootTx {
		txmatrices = predictor.profile.PeriodTree.Root.TxMatrix
		txmatrices = predictor.profile.RootTx
	} else if predictor.mode == PredictionModePhases {
		predictor.nextPhase()
		txmatrices = predictor.profile.Phases.Phases[predictor.currentPhase]
	} else if predictor.mode == PredictionModePeriods {
		predictor.nextPeriod(0)
		txmatrices = predictor.getCurrentPeriodTxMatrix()
	} else {
		fmt.Printf("warning: invalid prediction mode specified - falling back to root tx matrix")
		// fallback: root tx
		txmatrices = predictor.profile.RootTx
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
			//return nil, fmt.Errorf("error: %s (phase %d, periodPath %+v, txmatrix %+v)", err, predictor.currentPhase, predictor.periodPath, txmatrix)
			// plan b: take next step with highest stepProb
			//fmt.Printf("planb...\n")
			txstep, err = findStateByStateProbInTxmatrix(txmatrix)
			if err != nil {
				return nil, fmt.Errorf("error: %s (phase %d, periodPath %+v, txmatrix %+v)", err, predictor.currentPhase, predictor.periodPath, txmatrix)
			}
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

	return states, nil
}

func (predictor *Predictor) nextPhase() {
	currentPhase := predictor.currentPhase
	txmatrix := predictor.profile.Phases.Tx
	txstep, err := findStateHistoryInTxMatrix(txmatrix, fmt.Sprintf("%d", currentPhase))
	if err != nil {
		fmt.Printf("phase change error: %s\n", err)
		return
	}
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

func (predictor *Predictor) getCurrentPeriodTxMatrix() []models.TxMatrix {
	// select txmatrix according to periodPath and periodPathDepth
	if len(predictor.periodPath)-predictor.periodPathDepth < 0 {
		fmt.Printf("Warning: periodPathDepth too long for PeriodPath! Resizing to periodPathDepth %d\n", len(predictor.periodPath))
		predictor.periodPathDepth = len(predictor.periodPath)
	}

	path := predictor.periodPath[:predictor.periodPathDepth]
	fmt.Printf("path: %+v\n", path)
	node := predictor.profile.PeriodTree.GetNode(path)
	return node.TxMatrix
}

func (predictor *Predictor) nextPeriod(level int) bool {
	if len(predictor.periodPath) < predictor.periodPathDepth {
		fmt.Printf("periodPathDepth %d is larger than periodPath %d (%+v)! Impossible!", predictor.periodPathDepth, len(predictor.periodPath), predictor.periodPath)
		return false
	}

	moveOn := false
	nextLevel := level + 1
	if nextLevel < len(predictor.profile.Settings.PeriodSize) {
		// go down into tree
		nextLevel := level + 1
		moveOn = predictor.nextPeriod(nextLevel)
	} else {
		// at leaf node level already
		moveOn = (predictor.periodSizeCounter[level] >= predictor.profile.Settings.PeriodSize[level])
	}

	// check if running out of level
	moveOnUpperLevel := false
	if moveOn {
		predictor.periodPath[level] = predictor.periodPath[level] + 1
		if predictor.periodPath[level] >= predictor.profile.Settings.PeriodSize[level] {
			// reset position, start from 0  for current level
			predictor.periodPath[level] = 0
			moveOnUpperLevel = true
		}

		// reset for next node on tree level
		predictor.periodSizeCounter[level] = 0
	}

	// count for current level
	predictor.periodSizeCounter[level]++

	return moveOnUpperLevel

}

// SetState defines the given currentState for the next simulation
func (predictor *Predictor) SetState(currentState map[string]string) {
	predictor.currentState = currentState
}

// SetPhase defines the given phase for the next simulation
func (predictor *Predictor) SetPhase(currentPhase int) {
	predictor.currentPhase = currentPhase
}

// SetPeriodPath defines the current path in the period tree
func (predictor *Predictor) SetPeriodPath(periodPath []int, periodPathDepth int) {
	predictor.periodPath = periodPath
	predictor.periodPathDepth = periodPathDepth
}

// SetMode defines the given PredictionMode for the next simulation
func (predictor *Predictor) SetMode(mode PredictionMode) {
	predictor.mode = mode
}

// Simulate computes `steps` states using randomness and TSProfile's probabilities
func (predictor *Predictor) Simulate(steps int) ([][]models.TSState, error) {
	simulation := make([][]models.TSState, steps)
	for i := 0; i < steps; i++ {
		next, err := predictor.nextState()
		if err != nil {
			return nil, err
		}
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
	return simulation, nil
}

func (predictor *Predictor) initializeState() {
	var txmatrices []models.TxMatrix
	if predictor.mode == PredictionModeRootTx {
		txmatrices = predictor.profile.RootTx
	} else if predictor.mode == PredictionModePhases {
		txmatrices = predictor.profile.Phases.Phases[predictor.currentPhase]
	} else if predictor.mode == PredictionModePeriods {
		txmatrices = predictor.getCurrentPeriodTxMatrix()
	} else {
		fmt.Printf("warning: invalid prediction mode specified - falling back to root tx matrix")
		// fallback: root tx
		txmatrices = predictor.profile.RootTx
	}

	currentState := make(map[string]string)
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
