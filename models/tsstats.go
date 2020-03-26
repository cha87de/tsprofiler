package models

// TSStats contains default statistics
type TSStats struct {
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Stddev    float64 `json:"stddev"`
	Avg       float64 `json:"avg"`
	Count     int64   `json:"count"`
	StddevSum float64 `json:"stddevsum"`
}
