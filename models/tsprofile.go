package models

// TSProfile contains the resulting statistical profile
type TSProfile struct {
	Name     string            `json:"name"`
	Metrics  []TSProfileMetric `json:"metrics"`
	Settings Settings          `json:"settings"`
}
