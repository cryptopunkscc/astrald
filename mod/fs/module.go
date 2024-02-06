package fs

import (
	"github.com/cryptopunkscc/astrald/data"
	"path/filepath"
	"time"
)

const (
	ModuleName       = "fs"
	DBPrefix         = "fs__"
	MemorySetName    = ".fs.memory"
	ReadOnlySetName  = ".fs.readonly"
	ReadWriteSetName = ".fs.readwrite"
)

type Module interface {
	Find(id data.ID) []string
}

type EventFileChanged struct {
	Path      string
	OldID     data.ID
	NewID     data.ID
	IndexedAt time.Time
}

type EventFileAdded struct {
	Path   string
	DataID data.ID
}

type EventFileRemoved struct {
	Path   string
	DataID data.ID
}

type FileDescriptor struct {
	Paths []string
}

func (FileDescriptor) DescriptorType() string {
	return "mod.fs.file"
}
func (d FileDescriptor) String() string {
	if len(d.Paths) == 0 {
		return ""
	}
	return filepath.Base(d.Paths[0])
}
