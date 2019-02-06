package models

// NewPeriodTreeNode instantiates and returns a PeriodTreeNode with `size` children (recursively)
func NewPeriodTreeNode(size []int) PeriodTreeNode {
	maxChilds := size[0]
	children := make([]PeriodTreeNode, 0)
	if len(size) > 1 {
		// build children
		for i := 0; i < maxChilds; i++ {
			children = append(children, NewPeriodTreeNode(size[1:]))
		}
	}

	return PeriodTreeNode{
		MaxChilds: maxChilds,
		Children:  children,
		TxMatrix:  make([]TxMatrix, 0),
	}
}

// PeriodTreeNode describes a node holding a TxMatrix and (if not leaf node) children
type PeriodTreeNode struct {
	MaxChilds int              `json:"maxChilds"`
	Children  []PeriodTreeNode `json:"children"`
	TxMatrix  []TxMatrix       `json:"txmatrix"`
}

// GetNode returns the TreeNode which is located at `path`
func (periodTreeNode *PeriodTreeNode) GetNode(path []int) *PeriodTreeNode {
	currentPos := path[0]
	nextPos := path[1:]
	if len(nextPos) > 1 {
		return periodTreeNode.Children[currentPos].GetNode(nextPos)
	}
	return &periodTreeNode.Children[currentPos]
}
