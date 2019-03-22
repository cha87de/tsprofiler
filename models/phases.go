package models

// Phases holds a list of detected phases and a tx matrix describing the transitioning between these phases
type Phases struct {
	// A list of detected phases (outer array) for each metric (inner array)
	Phases [][]TxMatrix `json:"phases"`

	// Tx holds the transitions between the phases
	Tx TxMatrix `json:"tx"`
}
