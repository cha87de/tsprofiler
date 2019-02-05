package models

// NewPeriodTree instantiates and returns a new PeriodTree with the `size` child nodes
func NewPeriodTree(size []int) PeriodTree {
	return PeriodTree{
		Root: NewPeriodTreeNode(size),
	}
}

// PeriodTree contains the root of a PeriodTree
type PeriodTree struct {
	Root PeriodTreeNode `json:"root"`
}

func (periodTree *PeriodTree) GetNode(path []int) *PeriodTreeNode {
	return periodTree.Root.GetNode(path)
}
