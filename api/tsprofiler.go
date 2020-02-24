package api

import (
	models "github.com/cha87de/tsprofiler/models"
)

// TSProfiler defines the profilers interface to transfer time series data into
// a statistical profile
type TSProfiler interface {
	// Put allows applications to provide a new TSData input to the profiler
	Put(data models.TSInput)

	// Get generates an returns a profile based on previously put data
	Get() models.TSProfile

	GetCurrentStats() map[string]models.TSStats
	GetCurrentState() []models.TSState
	GetCurrentPhase() int

	// Terminate stops and removes the profiler
	Terminate()
}
