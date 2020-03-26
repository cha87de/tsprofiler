package models

import (
	"math/rand"
)

// NewPeriodTreeNode instantiates and returns a PeriodTreeNode with `size` children (recursively)
func NewPeriodTreeNode(size []int) PeriodTreeNode {
	maxChilds := 0
	maxCounts := 0
	children := make([]PeriodTreeNode, 0)
	if len(size) > 0 {
		if len(size) > 1 {
			// build children
			maxChilds = size[0]
			for i := 0; i < maxChilds; i++ {
				child := NewPeriodTreeNode(size[1:])
				maxCounts += child.MaxCounts
				children = append(children, child)
			}
		} else {
			maxCounts = size[0]
		}
	}
	return PeriodTreeNode{
		UUID:      rand.Intn(999),
		MaxChilds: maxChilds,
		Children:  children,
		MaxCounts: maxCounts,
		TxMatrix:  make([]TxMatrix, 0),
	}
}

// PeriodTreeNode describes a node holding a TxMatrix and (if not leaf node) children
type PeriodTreeNode struct {
	UUID      int
	MaxChilds int              `json:"maxChilds"`
	MaxCounts int              `json:"maxCounts"`
	Children  []PeriodTreeNode `json:"children"`
	TxMatrix  []TxMatrix       `json:"txmatrix"`
}

// GetNode returns the TreeNode which is located at `path`
func (periodTreeNode *PeriodTreeNode) GetNode(path []int) *PeriodTreeNode {
	if len(path) == 0 {
		return periodTreeNode
	}
	if len(path) > 1 {
		// go into child path
		return periodTreeNode.Children[path[0]].GetNode(path[1:])
	}
	if len(periodTreeNode.Children) > 0 {
		return &periodTreeNode.Children[path[0]]
	}
	return periodTreeNode
}
