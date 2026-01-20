package paths

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotAbsolute indicates that a path is relative when an absolute path was required.
var ErrNotAbsolute = errors.New("path is not absolute")

// PathUnderRoot checks if path is equal to root or a descendant of root.
func PathUnderRoot(path string, root string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)

	if path == root {
		return true
	}

	if root == "/" {
		return true
	}

	return strings.HasPrefix(path, root+string(filepath.Separator))
}

// WidestRoots filters the list to only roots that are not descendants of other roots.
func WidestRoots(roots []string) []string {
	var widest []string
	for _, root := range roots {
		underOther := false
		for _, other := range roots {
			if root != other && PathUnderRoot(root, other) {
				underOther = true
				break
			}
		}
		if !underOther {
			widest = append(widest, root)
		}
	}
	return widest
}

// PathTrie provides efficient coverage checks for filesystem subtrees.
type PathTrie struct {
	children  map[string]*PathTrie
	isEnd     bool // marks this node as a subtree root; all descendants are covered
	coversAll bool // set when filesystem root "/" is inserted
}

// NewPathTrie builds an immutable trie from absolute paths where each path covers itself and all descendants.
func NewPathTrie(paths []string) (*PathTrie, error) {
	t := &PathTrie{children: make(map[string]*PathTrie)}
	for _, path := range paths {
		if err := t.insert(path); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// splitAbsPath normalizes and splits an absolute path into segments, returning ErrNotAbsolute for relative paths.
func splitAbsPath(path string) ([]string, error) {
	if !filepath.IsAbs(path) {
		return nil, ErrNotAbsolute
	}
	return strings.Split(filepath.Clean(path), string(filepath.Separator)), nil
}

// insert adds path as a subtree root, returning ErrNotAbsolute for relative paths.
func (t *PathTrie) insert(path string) error {
	parts, err := splitAbsPath(path)
	if err != nil {
		return err
	}

	// Filesystem root "/" splits to ["", ""] after Clean
	isRoot := len(parts) == 0 || (len(parts) == 2 && parts[0] == "" && parts[1] == "")
	if isRoot {
		t.coversAll = true
		return nil
	}

	node := t
	for _, part := range parts {
		if part == "" {
			continue
		}
		if node.children[part] == nil {
			node.children[part] = &PathTrie{children: make(map[string]*PathTrie)}
		}
		node = node.children[part]
	}
	node.isEnd = true
	return nil
}

// Covers reports whether path falls under any registered subtree root, returning ErrNotAbsolute for relative paths.
func (t *PathTrie) Covers(path string) (bool, error) {
	if t.coversAll {
		return true, nil
	}

	parts, err := splitAbsPath(path)
	if err != nil {
		return false, err
	}

	node := t
	for _, part := range parts {
		if part == "" {
			continue
		}
		if node.isEnd {
			return true, nil
		}
		node = node.children[part]
		if node == nil {
			return false, nil
		}
	}
	return node.isEnd, nil
}

// WalkDir walks the directory tree from root using breadth-first traversal,
// calling fn for each regular file. Errors reading directories are skipped.
// This avoids keeping many directory handles open, unlike filepath.WalkDir.
func WalkDir(ctx context.Context, root string, fn func(path string) error) error {
	queue := []string{root}

	for len(queue) > 0 {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		dir := queue[0]
		queue = queue[1:]

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				queue = append(queue, path)
			} else if entry.Type().IsRegular() {
				if err := fn(path); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
