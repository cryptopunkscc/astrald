package fs

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "fs"
const AllSetName = "mod.fs.all"
const MemorySetName = "mod.fs.mem"
const ReadOnlySetName = "mod.fs.ro"
const ReadWriteSetName = "mod.fs.rw"

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

func (FileDescriptor) InfoType() string {
	return "mod.fs.file"
}
