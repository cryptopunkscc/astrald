package archives

import "github.com/cryptopunkscc/astrald/object"

type ArchiveDesc struct {
	Files []ArchiveEntry
}

func (ArchiveDesc) Type() string {
	return "mod.archives.archive"
}

type ArchiveEntry struct {
	ObjectID object.ID
	Path     string
}

type EntryDesc struct {
	Containers []Container
}

func (EntryDesc) Type() string {
	return "mod.archives.entry"
}

type Container struct {
	ObjectID object.ID
	Path     string
}
