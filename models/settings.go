package models

import (
	"time"
)

// Settings defines settings for TSProfiler
type Settings struct {
	// BufferSize defines the amount of TSData items before a new state is transitioned
	BufferSize int `json:"buffersize"`

	// Name allows to identify the profiler, e.g. for human readable differentiation
	Name string `json:"-"`

	// States defines the amount of states to discretize the measurements
	States int `json:"states"`

	// History defines the amount of previous, historic state changes to be considered
	History int `json:"history"`

	// FilterStdDevs defines the amount of stddevs which are max. allowed for data items before skipped as outliers
	FilterStdDevs int `json:"filterstddevs"`

	// FixBound defines if min/max are fixed or dynamic depending on occurred values
	FixBound bool `json:"fixbound"`

	// OutputFreq controls the frequency in which the profiler calls the OutputCallback function (if not set, profile has to be retrieved manually)
	OutputFreq time.Duration `json:"-"`

	// OutputCallback defines the callback function for `TSProfile`s every `OutputFreq`
	OutputCallback func(data TSProfile) `json:"-"`

	// PeriodSize defines the amount and size of periods
	PeriodSize []int `json:"periodsize"`

	// Phase Change Detection settings (likeliness over history)
	PhaseChangeLikeliness float32 `json:"phaseChangeLikeliness"`
	// Phase Change Detection settings (state history length)
	PhaseChangeHistory int64 `json:"phaseChangeHistory"`
	// Phase Change Detection settings (state history fade out)
	PhaseChangeHistoryFadeout bool `json:"phaseChangeHistoryFadeout"`
}
