package models

// TSState describes a single discretized state
type TSState struct {
	Metric     string
	Statistics TSStats
	State      State
}
