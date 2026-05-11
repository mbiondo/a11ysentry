package domain

// FileNode represents a hierarchical node in a source tree (e.g., Layout wrapping a Page).
type FileNode struct {
	FilePath     string
	Children     []*FileNode
	IsCycle      bool   // true if this node points back to an ancestor, preventing infinite recursion
	IsOpaque     bool   // true if internal source is not available
	OpaqueSource string // Identifier of the external source (e.g. "@mui/material")
}

// Flatten returns a flat slice of all file paths in the tree.
func (n *FileNode) Flatten() []string {
	if n == nil {
		return nil
	}
	res := []string{n.FilePath}
	for _, c := range n.Children {
		res = append(res, c.Flatten()...)
	}
	return res
}

