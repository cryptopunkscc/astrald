package fs

import (
	"io/fs"
	"time"
)

var _ fs.FileInfo = FileInfo{}

type FileInfo struct {
	*File
}

func (f FileInfo) Name() string {
	return f.ID.String()
}

func (f FileInfo) Size() int64 {
	return int64(f.ID.Size)
}

func (f FileInfo) Mode() fs.FileMode {
	return 0444
}

func (f FileInfo) ModTime() time.Time {
	return time.Unix(0, 0)
}

func (f FileInfo) IsDir() bool {
	return false
}

func (f FileInfo) Sys() any {
	return nil
}
