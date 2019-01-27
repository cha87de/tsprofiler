package models

// TXStep expresses a single state in a markov chain / transition matrix
type TXStep struct {
	NextStateProbs []int `json:"nextProbs"`
	StepProb       int   `json:"probability"`
}
