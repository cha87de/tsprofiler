package models

// TSProfile contains the resulting statistical profile
type TSProfile struct {
	Name       string     `json:"name"`
	PeriodTree PeriodTree `json:"periodTree"`
	Phases     Phases     `json:"phases"`
	Settings   Settings   `json:"settings"`
}
