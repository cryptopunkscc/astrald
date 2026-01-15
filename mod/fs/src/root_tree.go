package fs

import (
	"path/filepath"
	"sort"
	"strings"
)

// RootTree maintains a hierarchy of root paths.
// When a parent root is added, child roots are consolidated under it.
type RootTree struct {
	roots map[string]*rootNode
}

type rootNode struct {
	path     string
	interest int // reference count for this specific root
	parent   *rootNode
	children map[string]*rootNode
}

func NewRootTree() *RootTree {
	return &RootTree{
		roots: make(map[string]*rootNode),
	}
}

// Add registers a root and restructures the hierarchy.
func (t *RootTree) Add(root string) {
	if _, exists := t.roots[root]; exists {
		return
	}

	node := &rootNode{
		path:     root,
		children: make(map[string]*rootNode),
	}

	var parent *rootNode
	var childrenToAdopt []*rootNode

	for existingPath, existingNode := range t.roots {
		if Contains(existingPath, root) {
			if parent == nil || len(existingPath) > len(parent.path) {
				parent = existingNode
			}
			continue
		}

		if Contains(root, existingPath) {
			if existingNode.parent != nil && Contains(root, existingNode.parent.path) {
				continue
			}
			childrenToAdopt = append(childrenToAdopt, existingNode)
		}
	}

	node.parent = parent
	if parent != nil {
		parent.children[root] = node
	}

	for _, child := range childrenToAdopt {
		if child.parent != nil {
			delete(child.parent.children, child.path)
		}
		child.parent = node
		node.children[child.path] = child
	}

	t.roots[root] = node
}

// Remove unregisters a root. Children are promoted to parent level.
func (t *RootTree) Remove(root string) {
	node, exists := t.roots[root]
	if !exists {
		return
	}

	for _, child := range node.children {
		child.parent = node.parent
		if node.parent != nil {
			node.parent.children[child.path] = child
		}
	}

	if node.parent != nil {
		delete(node.parent.children, root)
	}

	delete(t.roots, root)
}

// Has returns true if root is registered.
func (t *RootTree) Has(root string) bool {
	_, ok := t.roots[root]
	return ok
}

// FindWidest returns the widest (most parent) root that contains path.
func (t *RootTree) FindWidest(path string) string {
	var widest string
	var widestLen int

	for rootPath := range t.roots {
		if Contains(rootPath, path) {
			if widest == "" || len(rootPath) < widestLen {
				widest = rootPath
				widestLen = len(rootPath)
			}
		}
	}

	return widest
}

// FindAll returns all roots that contain path, sorted widest-first.
func (t *RootTree) FindAll(path string) []string {
	var result []string

	for rootPath := range t.roots {
		if Contains(rootPath, path) {
			result = append(result, rootPath)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return len(result[i]) < len(result[j])
	})

	return result
}

// TopLevel returns all roots with no parent.
func (t *RootTree) TopLevel() []string {
	var result []string

	for rootPath, node := range t.roots {
		if node.parent == nil {
			result = append(result, rootPath)
		}
	}

	sort.Strings(result)
	return result
}

// Children returns immediate child roots of the given root.
func (t *RootTree) Children(root string) []string {
	node, exists := t.roots[root]
	if !exists {
		return nil
	}

	var result []string
	for childPath := range node.children {
		result = append(result, childPath)
	}

	sort.Strings(result)
	return result
}

// Parent returns the parent root. Empty string if no parent.
func (t *RootTree) Parent(root string) string {
	node, exists := t.roots[root]
	if !exists || node.parent == nil {
		return ""
	}
	return node.parent.path
}

// Len returns total number of registered roots.
func (t *RootTree) Len() int {
	return len(t.roots)
}

// All returns all registered roots in sorted order.
func (t *RootTree) All() []string {
	result := make([]string, 0, len(t.roots))
	for root := range t.roots {
		result = append(result, root)
	}
	sort.Strings(result)
	return result
}

// Contains returns true if path is within root's directory tree.
func Contains(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	// Path is outside root if relative path starts with ".."
	if strings.HasPrefix(rel, "..") {
		return false
	}

	// Path equals root if relative path is "."
	if rel == "." {
		return false
	}

	return true
}

// SetInterest sets the interest count for a specific root.
func (t *RootTree) SetInterest(root string, interest int) {
	node, exists := t.roots[root]
	if exists {
		node.interest = interest
	}
}

// GetInterest returns the interest count for a specific root.
func (t *RootTree) GetInterest(root string) int {
	node, exists := t.roots[root]
	if !exists {
		return 0
	}
	return node.interest
}

// GetAggregatedInterest returns total interest for root and all its descendants.
func (t *RootTree) GetAggregatedInterest(root string) int {
	node, exists := t.roots[root]
	if !exists {
		return 0
	}

	total := node.interest
	for _, child := range node.children {
		total += t.GetAggregatedInterest(child.path)
	}
	return total
}
