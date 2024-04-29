package zip

import (
	"archive/zip"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"io/fs"
)

var _ io.ReadSeeker = &contentReader{}
var _ objects.Reader = &contentReader{}

type contentReader struct {
	zip      *zip.Reader
	path     string
	objectID object.ID

	file fs.File
	pos  int64
}

func (r *contentReader) Read(p []byte) (n int, err error) {
	n, err = r.file.Read(p)
	r.pos += int64(n)
	return n, err
}

func (r *contentReader) Seek(offset int64, whence int) (int64, error) {
	var target int64

	switch whence {
	case io.SeekStart:
		target = offset
	case io.SeekCurrent:
		target = int64(r.pos) + offset
	case io.SeekEnd:
		target = int64(r.objectID.Size) + offset
	}

	if target == r.pos {
		return r.pos, nil
	}

	if target > r.pos {
		err := streams.Skip(r, uint64(target-r.pos))
		return r.pos, err
	}

	var err = r.open()
	if err != nil {
		return 0, err
	}

	err = streams.Skip(r, uint64(target))
	return r.pos, err
}

func (r *contentReader) Close() (err error) {
	if r.file != nil {
		err = r.file.Close()
	}
	r.file = nil
	return
}

func (r *contentReader) open() (err error) {
	if r.file != nil {
		r.file.Close()
		r.file = nil
		r.pos = 0
	}

	r.file, err = r.zip.Open(r.path)

	return
}

func (r *contentReader) Info() *objects.ReaderInfo {
	return &objects.ReaderInfo{Name: "mod.zip"}
}
