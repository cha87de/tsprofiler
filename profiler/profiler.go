package profiler

import (
	"time"

	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/profiler/buffer"
	"github.com/cha87de/tsprofiler/profiler/counter"
	"github.com/cha87de/tsprofiler/profiler/discretizer"
)

// NewProfiler creates and returns a new TSProfiler, configured with given Settings
func NewProfiler(settings models.Settings) *Profiler {
	profiler := profiler{}
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
	counter     counter.Counter
}

func (profiler *Profiler) initialize(settings models.Settings) {
	profiler.input = make(chan models.TSInput, 0)
	profiler.settings = settings
	profiler.stopped = false

	// initialize sub components
	profiler.buffer = buffer.NewBuffer(settings.FilterStdDevs, profiler)
	profiler.discretizer = discretizer.NewDiscretizer(settings.States)
	profiler.counter = counter.NewCounter()

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

func (profiler *Profiler) GetCurrentState() []models.TSState {
	return BULLSHIT
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

		if itemCount > profiler.settings.BufferSize {
			// buffer is full, trigger discretizer!
			tsbuffers := profiler.buffer.Reset()
			tsstates := profiler.discretizer.Discretize(tsbuffers)
			profiler.counters.Count(tsstates)
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
func (profiler *profiler) generateProfile() models.TSProfile {
	var metrics []models.TSProfileMetric
	profiler.metricsAccess.Lock()
	defer profiler.metricsAccess.Unlock()
	for _, metricProfiler := range profiler.metrics {
		maxCount := float64(metricProfiler.counts.stats.Count) / float64(profiler.settings.BufferSize)
		txmatrix := computeProbabilities(metricProfiler.counts.stateChangeCounter, maxCount)
		// fmt.Printf("counter %+v, probs: %+v\n", metricProfiler.counts.stateChangeCounter, txmatrix)
		metrics = append(metrics, models.TSProfileMetric{
			Name:     metricProfiler.name,
			TXMatrix: txmatrix,
			Stats:    metricProfiler.counts.stats,
		})
	}
	return models.TSProfile{
		Name:     profiler.settings.Name,
		Metrics:  metrics,
		Settings: profiler.settings,
	}
}