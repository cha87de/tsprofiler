package models

// TSProfileMetric describes for one metric a statistical profile
type TSProfileMetric struct {
	Name     string            `json:"name"`
	TXMatrix map[string]TXStep `json:"txmatrix"`
	Stats    TSStats           `json:"stats"`
}
