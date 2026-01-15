package fs

import (
	"testing"

	"github.com/cryptopunkscc/astrald/sig"
)

func TestIsUnderActiveRoot(t *testing.T) {
	tests := []struct {
		name        string
		roots       []string
		path        string
		expectMatch bool
	}{
		{
			name:        "exact match on root",
			roots:       []string{"/workspace"},
			path:        "/workspace",
			expectMatch: true,
		},
		{
			name:        "path directly under root",
			roots:       []string{"/workspace"},
			path:        "/workspace/product",
			expectMatch: true,
		},
		{
			name:        "deeply nested path under root",
			roots:       []string{"/workspace/product"},
			path:        "/workspace/product/app/server.go",
			expectMatch: true,
		},
		{
			name:        "path with similar prefix but different directory",
			roots:       []string{"/workspace/product"},
			path:        "/workspace/product-tools/build.sh",
			expectMatch: false,
		},
		{
			name:        "path under one of multiple roots",
			roots:       []string{"/workspace/product", "/workspace/shared", "/workspace"},
			path:        "/workspace/shared/lib/util.go",
			expectMatch: true,
		},
		{
			name:        "path not under any root",
			roots:       []string{"/workspace/product", "/workspace/shared"},
			path:        "/home/user/file.txt",
			expectMatch: false,
		},
		{
			name:        "empty roots set",
			roots:       []string{},
			path:        "/workspace/product/app/server.go",
			expectMatch: false,
		},
		{
			name:        "path matches parent root when child removed",
			roots:       []string{"/workspace"},
			path:        "/workspace/product/app/server.go",
			expectMatch: true,
		},
		{
			name:        "path with trailing slash in root",
			roots:       []string{"/workspace/product"},
			path:        "/workspace/product/app",
			expectMatch: true,
		},
		{
			name:        "similar prefix at character boundary",
			roots:       []string{"/work"},
			path:        "/workspace/file.txt",
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi := &FileIndexer{
				activeRoots: sig.Set[string]{},
			}

			for _, root := range tt.roots {
				fi.activeRoots.Add(root)
			}

			result := fi.IsUnderActiveRoot(tt.path)
			if result != tt.expectMatch {
				t.Errorf("IsUnderActiveRoot(%q) with roots %v = %v, want %v",
					tt.path, tt.roots, result, tt.expectMatch)
			}
		})
	}
}

func TestIsUnderActiveRoot_OverlappingRoots(t *testing.T) {
	fi := &FileIndexer{
		activeRoots: sig.Set[string]{},
	}

	// Add overlapping roots
	fi.activeRoots.Add("/workspace")
	fi.activeRoots.Add("/workspace/product")
	fi.activeRoots.Add("/workspace/shared")

	tests := []struct {
		path        string
		expectMatch bool
	}{
		{"/workspace/product/app/server.go", true},
		{"/workspace/shared/lib/util.go", true},
		{"/workspace/infra/deploy.yaml", true},      // matched by parent /workspace
		{"/workspace/product-tools/build.sh", true}, // matched by parent /workspace
		{"/home/user/file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := fi.IsUnderActiveRoot(tt.path)
			if result != tt.expectMatch {
				t.Errorf("IsUnderActiveRoot(%q) = %v, want %v",
					tt.path, result, tt.expectMatch)
			}
		})
	}
}

func TestIsUnderActiveRoot_RootRemoval(t *testing.T) {
	fi := &FileIndexer{
		activeRoots: sig.Set[string]{},
	}

	// Start with overlapping roots
	fi.activeRoots.Add("/workspace")
	fi.activeRoots.Add("/workspace/product")

	path := "/workspace/product/app/server.go"
	if !fi.IsUnderActiveRoot(path) {
		t.Errorf("Expected path to be under active root initially")
	}

	// Remove child root
	fi.activeRoots.Remove("/workspace/product")
	if !fi.IsUnderActiveRoot(path) {
		t.Errorf("Expected path to still match parent root /workspace")
	}

	// Remove parent root
	fi.activeRoots.Remove("/workspace")
	if fi.IsUnderActiveRoot(path) {
		t.Errorf("Expected path to not match after removing all roots")
	}
}
