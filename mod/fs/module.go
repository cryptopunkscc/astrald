package fs

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "fs"
const IndexNameAll = "mod.fs.all"

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

const FileDescriptorType = "mod.fs.file"

type FileDescriptor struct {
	Paths []string
}
