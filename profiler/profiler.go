package profiler

import (
	"sync"
	"time"

	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/profiler/buffer"
	"github.com/cha87de/tsprofiler/profiler/counter"
	"github.com/cha87de/tsprofiler/profiler/discretizer"
	"github.com/cha87de/tsprofiler/profiler/period"
	"github.com/cha87de/tsprofiler/profiler/phase"
)

// NewProfiler creates and returns a new TSProfiler, configured with given Settings
func NewProfiler(settings models.Settings) *Profiler {
	profiler := Profiler{}
	profiler.initialize(settings)
	return &profiler
}

// Profiler is the TSProfiler implementation of spec.TSProfiler
type Profiler struct {
	input    chan models.TSInput
	settings models.Settings
	stopped  bool

	// state
	overallCounter counter.Counter
	lastStates     []models.TSState

	access *sync.Mutex

	// sub components
	buffer      buffer.Buffer
	discretizer discretizer.Discretizer
	period      period.Period
	phase       phase.Phase
}

func (profiler *Profiler) initialize(settings models.Settings) {
	profiler.input = make(chan models.TSInput, 0)
	profiler.settings = settings
	profiler.stopped = false

	// initialize sub components
	profiler.buffer = buffer.NewBuffer(settings.FilterStdDevs, profiler)
	profiler.discretizer = discretizer.NewDiscretizer(settings.States, settings.FixBound, profiler)
	profiler.period = period.NewPeriod(settings.History, settings.States, settings.BufferSize, settings.PeriodSize, profiler)
	profiler.phase = phase.NewPhase(settings.History, settings.States, settings.BufferSize, settings.PhaseChangeLikeliness, settings.PhaseChangeMincount, profiler)

	// initialize root tx counter
	profiler.overallCounter = counter.NewCounter(settings.History, settings.States, settings.BufferSize, profiler)
	profiler.lastStates = make([]models.TSState, 0)
	profiler.access = &sync.Mutex{}

	// start input & output background routines
	go profiler.outputRunner()
	go profiler.inputListener()
}

// Put adds a TSData item to the profiler
func (profiler *Profiler) Put(data models.TSInput) {
	profiler.input <- data
}

// Get generates an returns a profile based on previously put data
func (profiler *Profiler) Get() models.TSProfile {
	return profiler.generateProfile()
}

// GetCurrentStats returns the current stats for each metric
func (profiler *Profiler) GetCurrentStats() map[string]models.TSStats {
	return profiler.overallCounter.GetStats()
}

// GetCurrentState returns the current state for each metric
func (profiler *Profiler) GetCurrentState() []models.TSState {
	return profiler.lastStates
}

// GetCurrentPhase returns the current phase id
func (profiler *Profiler) GetCurrentPhase() int {
	return profiler.phase.GetPhase()
}

// GetCurrentPeriodPath returns the current period path
func (profiler *Profiler) GetCurrentPeriodPath() []int {
	return profiler.period.GetCurrentPeriodPath()
}

// Terminate stops and removes the profiler
func (profiler *Profiler) Terminate() {
	profiler.stopped = true
	close(profiler.input)
}

// inputListener handles incoming tsdata item from input channel
func (profiler *Profiler) inputListener() {
	itemCount := 0
	for !profiler.stopped {
		input := <-profiler.input

		profiler.buffer.Add(input)
		itemCount++

		if itemCount >= profiler.settings.BufferSize {
			// buffer is full, trigger discretizer!
			tsbuffers := profiler.buffer.Reset()
			tsstates := profiler.discretizer.Discretize(tsbuffers)

			profiler.access.Lock()

			// global all time counting
			profiler.overallCounter.Count(tsstates)

			// update lastState
			profiler.lastStates = tsstates

			// call sub components
			profiler.period.Count(tsstates)
			profiler.phase.Count(tsstates)

			itemCount = 0

			profiler.access.Unlock()
		}
	}
}

// outputRunner schedules periodic tsprofile generation (if OutputFreq && OutputCallback are set)
func (profiler *Profiler) outputRunner() {
	if profiler.settings.OutputCallback == nil || profiler.settings.OutputFreq == 0 {
		// no automated output specified
		return
	}
	for !profiler.stopped {
		start := time.Now()
		profile := profiler.generateProfile()
		profiler.settings.OutputCallback(profile)
		nextRun := start.Add(profiler.settings.OutputFreq)
		time.Sleep(nextRun.Sub(time.Now()))
	}
}

// generateProfile collects the necessary data to return a TSProfile
func (profiler *Profiler) generateProfile() models.TSProfile {
	periodTree := profiler.period.GetTx()
	//periodTree.Root.TxMatrix = profiler.overallCounter.GetTx()
	rootTx := profiler.overallCounter.GetTx()
	phases := profiler.phase.GetPhasesTx()
	return models.TSProfile{
		Name:       profiler.settings.Name,
		RootTx:     rootTx,
		PeriodTree: periodTree,
		Phases:     phases,
		Settings:   profiler.settings,
	}
}
