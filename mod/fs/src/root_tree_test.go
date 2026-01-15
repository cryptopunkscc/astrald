package fs

import (
	"reflect"
	"testing"
)

func TestRootTree_Basic(t *testing.T) {
	rt := NewRootTree()

	if rt.Len() != 0 {
		t.Errorf("Len() = %d, want 0", rt.Len())
	}

	rt.Add("/home/user/music")
	rt.Add("/opt/backup")

	if rt.Len() != 2 {
		t.Errorf("Len() = %d, want 2", rt.Len())
	}

	if !rt.Has("/home/user/music") {
		t.Error("Has(/home/user/music) = false, want true")
	}

	rt.Remove("/opt/backup")
	if rt.Has("/opt/backup") {
		t.Error("Has(/opt/backup) = true after remove, want false")
	}
}

func TestRootTree_Hierarchy(t *testing.T) {
	rt := NewRootTree()

	// Add specific subdirectory first
	rt.Add("/home/user/projects/work/repo")
	topLevel := rt.TopLevel()
	if len(topLevel) != 1 || topLevel[0] != "/home/user/projects/work/repo" {
		t.Errorf("TopLevel() = %v, want [/home/user/projects/work/repo]", topLevel)
	}

	// Add parent directory - should restructure
	rt.Add("/home/user/projects")
	topLevel = rt.TopLevel()
	if len(topLevel) != 1 || topLevel[0] != "/home/user/projects" {
		t.Errorf("TopLevel() after adding parent = %v, want [/home/user/projects]", topLevel)
	}

	// Verify repo is now child of projects
	parent := rt.Parent("/home/user/projects/work/repo")
	if parent != "/home/user/projects" {
		t.Errorf("Parent() = %q, want /home/user/projects", parent)
	}

	children := rt.Children("/home/user/projects")
	if len(children) != 1 || children[0] != "/home/user/projects/work/repo" {
		t.Errorf("Children() = %v, want [/home/user/projects/work/repo]", children)
	}
}

func TestRootTree_FindWidest(t *testing.T) {
	rt := NewRootTree()
	rt.Add("/home/user")
	rt.Add("/home/user/documents")
	rt.Add("/home/user/documents/notes")

	widest := rt.FindWidest("/home/user/documents/notes/file.txt")
	if widest != "/home/user" {
		t.Errorf("FindWidest() = %q, want /home/user", widest)
	}

	// Non-matching path
	widest = rt.FindWidest("/tmp/random.txt")
	if widest != "" {
		t.Errorf("FindWidest() for non-matching = %q, want empty", widest)
	}
}

func TestRootTree_FindAll(t *testing.T) {
	rt := NewRootTree()
	rt.Add("/home/user")
	rt.Add("/home/user/documents")

	all := rt.FindAll("/home/user/documents/file.txt")
	expected := []string{"/home/user", "/home/user/documents"}

	if !reflect.DeepEqual(all, expected) {
		t.Errorf("FindAll() = %v, want %v", all, expected)
	}
}

func TestRootTree_Siblings(t *testing.T) {
	rt := NewRootTree()
	rt.Add("/home/user/documents")
	rt.Add("/home/user/projects")

	// Both should be top-level (no common parent yet)
	topLevel := rt.TopLevel()
	if len(topLevel) != 2 {
		t.Errorf("TopLevel() = %v, want 2 roots", topLevel)
	}

	// Add common parent
	rt.Add("/home/user")
	topLevel = rt.TopLevel()
	if len(topLevel) != 1 || topLevel[0] != "/home/user" {
		t.Errorf("TopLevel() after parent = %v, want [/home/user]", topLevel)
	}

	// Both should be children
	children := rt.Children("/home/user")
	expected := []string{"/home/user/documents", "/home/user/projects"}
	if !reflect.DeepEqual(children, expected) {
		t.Errorf("Children() = %v, want %v", children, expected)
	}
}

func TestRootTree_RemovePromotesChildren(t *testing.T) {
	rt := NewRootTree()
	rt.Add("/home/user")
	rt.Add("/home/user/documents")
	rt.Add("/home/user/projects")

	// Remove parent
	rt.Remove("/home/user")

	// Children should be promoted to top-level
	topLevel := rt.TopLevel()
	expected := []string{"/home/user/documents", "/home/user/projects"}
	if !reflect.DeepEqual(topLevel, expected) {
		t.Errorf("TopLevel() after remove = %v, want %v", topLevel, expected)
	}

	// Children should have no parent
	if rt.Parent("/home/user/documents") != "" {
		t.Error("documents should have no parent after promotion")
	}
}

func TestRootTree_MultiLevel(t *testing.T) {
	rt := NewRootTree()

	// Deep hierarchy
	rt.Add("/home/user/projects/work/repo1")
	rt.Add("/home/user/projects/work")
	rt.Add("/home/user/projects")
	rt.Add("/home/user")

	// Should track only at widest level
	topLevel := rt.TopLevel()
	if len(topLevel) != 1 || topLevel[0] != "/home/user" {
		t.Errorf("TopLevel() = %v, want [/home/user]", topLevel)
	}

	// Verify full hierarchy
	if rt.Parent("/home/user/projects") != "/home/user" {
		t.Error("projects parent should be /home/user")
	}
	if rt.Parent("/home/user/projects/work") != "/home/user/projects" {
		t.Error("work parent should be projects")
	}
	if rt.Parent("/home/user/projects/work/repo1") != "/home/user/projects/work" {
		t.Error("repo1 parent should be work")
	}

	// Deep file should find widest root
	widest := rt.FindWidest("/home/user/projects/work/repo1/src/main.go")
	if widest != "/home/user" {
		t.Errorf("FindWidest() = %q, want /home/user", widest)
	}
}

func TestRootTree_All(t *testing.T) {
	rt := NewRootTree()
	rt.Add("/home/user/music")
	rt.Add("/opt/backup")
	rt.Add("/home/user/music/albums")

	all := rt.All()
	expected := []string{"/home/user/music", "/home/user/music/albums", "/opt/backup"}

	if !reflect.DeepEqual(all, expected) {
		t.Errorf("All() = %v, want %v", all, expected)
	}
}
