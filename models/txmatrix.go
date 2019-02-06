package models

// TxMatrix describes for one metric a statistical profile
type TxMatrix struct {
	Metric      string            `json:"metric"`
	Transitions map[string]TXStep `json:"transitions"`
	Stats       TSStats           `json:"stats"`
}
