package fs

import (
	"path/filepath"
)

const (
	ModuleName = "fs"
	DBPrefix   = "fs__"
)

type Module interface {
}

type FileDesc struct {
	Paths []string
}

func (FileDesc) Type() string {
	return "mod.fs.file"
}
func (d FileDesc) String() string {
	if len(d.Paths) == 0 {
		return ""
	}
	return filepath.Base(d.Paths[0])
}
