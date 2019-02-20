package profiler

import (
	"time"

	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/profiler/buffer"
	"github.com/cha87de/tsprofiler/profiler/discretizer"
	"github.com/cha87de/tsprofiler/profiler/period"
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

	// sub components
	buffer      buffer.Buffer
	discretizer discretizer.Discretizer
	period      period.Period
}

func (profiler *Profiler) initialize(settings models.Settings) {
	profiler.input = make(chan models.TSInput, 0)
	profiler.settings = settings
	profiler.stopped = false

	// initialize sub components
	profiler.buffer = buffer.NewBuffer(settings.FilterStdDevs, profiler)
	profiler.discretizer = discretizer.NewDiscretizer(settings.States, settings.FixBound, profiler)
	profiler.period = period.NewPeriod(settings.History, settings.States, settings.BufferSize, settings.PeriodSize, profiler)

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

// GetCurrentState returns the current state for each metric
func (profiler *Profiler) GetCurrentState() map[string]models.TSStats {
	return profiler.period.GetStats()
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
			profiler.period.Count(tsstates)
			itemCount = 0
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
	return models.TSProfile{
		Name:       profiler.settings.Name,
		PeriodTree: periodTree,
		Settings:   profiler.settings,
	}
}
