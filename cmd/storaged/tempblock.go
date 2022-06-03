package main

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/proto/block"
	"os"
	"path/filepath"
)

var _ block.Block = &TempBlock{}

type TempBlock struct {
	File *os.File

	Path string
}

func (b TempBlock) Read(p []byte) (n int, err error) {
	return b.File.Read(p)
}

func (b TempBlock) Write(p []byte) (n int, err error) {
	return b.File.Write(p)
}

func (b TempBlock) Seek(offset int64, whence int) (int64, error) {
	return b.File.Seek(offset, whence)
}

func (b TempBlock) Finalize() (data.ID, error) {
	b.File.Close()

	nf, err := os.Open(b.Path)
	if err != nil {
		return data.ID{}, err
	}

	resolvedID, err := data.ResolveAll(nf)
	if err != nil {
		return data.ID{}, err
	}

	resolvedPath := filepath.Join(filepath.Dir(b.Path), resolvedID.String())

	if err := os.Rename(b.Path, resolvedPath); err != nil {
		return data.ID{}, err
	}

	return resolvedID, nil
}

func (b TempBlock) End() error {
	return nil
}
