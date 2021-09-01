package fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"io"
	"log"
	"syscall"
)

type fileSender struct {
	fs.Inode
	writer io.WriteCloser
	name   string
}

var _ fs.FileWriter = new(fileSender)
var _ fs.FileFlusher = new(fileSender)

func (f *fileSender) Write(
	ctx context.Context,
	data []byte,
	off int64,
) (
	written uint32,
	errno syscall.Errno,
) {
	l, err := f.writer.Write(data)
	if err != nil {
		log.Println("cannot write", err)
		return uint32(l), syscall.EPIPE
	}
	return uint32(l), syscall.F_OK
}

func (f *fileSender) Flush(context.Context) syscall.Errno {
	err := f.writer.Close()
	if err != nil {
		log.Println("cannot flush", err)
		return syscall.EPIPE
	}
	return syscall.F_OK
}
