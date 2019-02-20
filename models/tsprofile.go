package models

// TSProfile contains the resulting statistical profile
type TSProfile struct {
	Name       string     `json:"name"`
	PeriodTree PeriodTree `json:"periodTree"`
	Settings   Settings   `json:"settings"`
}
