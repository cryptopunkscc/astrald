package fs

import (
	"path/filepath"
	"strings"

	"github.com/cryptopunkscc/astrald/mod/fs"
)

func pathUnderRoot(path string, root string) bool {
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

func widestRoots(roots []string) []string {
	var widest []string
	for _, root := range roots {
		underOther := false
		for _, other := range roots {
			if root != other && pathUnderRoot(root, other) {
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

// pathTrie provides efficient coverage checks for filesystem subtrees.
type pathTrie struct {
	children  map[string]*pathTrie
	isEnd     bool // marks this node as a subtree root; all descendants are covered
	coversAll bool // set when filesystem root "/" is inserted
}

// newPathTrie builds an immutable trie from absolute paths.
// Each path becomes a subtree root covering itself and all descendants.
func newPathTrie(paths []string) (*pathTrie, error) {
	t := &pathTrie{children: make(map[string]*pathTrie)}
	for _, path := range paths {
		if err := t.insert(path); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// splitAbsPath normalizes and splits an absolute path into segments.
// Returns fs.ErrNotAbsolute if path is not absolute.
func splitAbsPath(path string) ([]string, error) {
	if !filepath.IsAbs(path) {
		return nil, fs.ErrNotAbsolute
	}
	return strings.Split(filepath.Clean(path), string(filepath.Separator)), nil
}

// insert adds path as a subtree root. Do not call after construction.
func (t *pathTrie) insert(path string) error {
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
			node.children[part] = &pathTrie{children: make(map[string]*pathTrie)}
		}
		node = node.children[part]
	}
	node.isEnd = true
	return nil
}

// covers reports whether path falls under any registered subtree root.
// Returns fs.ErrNotAbsolute if path is not absolute.
func (t *pathTrie) covers(path string) (bool, error) {
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
