package service

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/file"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/remote"
	"github.com/mitchellh/ioprogress"
	"io"
	"time"
)

var _ api.FileService = File{}

type File api.Core

func (c File) Info(uri string) (files []api.Info, err error) {
	return c.resolve().Info(uri)
}

func (c File) CopyFrom(reader io.Reader, offer *api.Offer) (err error) {
	for _, info := range offer.Files {
		if err = c.copyFileFrom(reader, offer, info); err != nil {
			return
		}
	}
	return
}

func (c File) copyFileFrom(reader io.Reader, offer *api.Offer, info api.Info) (err error) {
	// Obtain writer
	storage := file.Storage(c)
	if info.IsDir {
		err := storage.MkDir(info.Path, info.Perm)
		if err != nil && !storage.IsExist(err) {
			c.Println("Cannot make dir", info.Path, err)
			return err
		}
	} else {
		writer, err := storage.FileWriter(info.Path, info.Perm)
		if err != nil {
			c.Println("Cannot get writer for", info.Path, err)
			return err
		}
		defer writer.Close()
		// Copy bytes
		progress := &ioprogress.Reader{
			Reader:       reader,
			Size:         info.Size,
			DrawInterval: 200 * time.Millisecond,
			DrawFunc: func(progress int64, size int64) error {
				status := fmt.Sprintf("download: %s %d/%dB", info.Path, progress, size)
				Incoming(api.Core(c)).Update(offer, status, false)
				return nil
			},
		}
		_, err = io.CopyN(writer, progress, info.Size)
		if err != nil {
			c.Println("Cannot copy", info.Path, err)
			return err
		}
		err = writer.Close()
		if err != nil {
			c.Println("Cannot close info", info.Path, err)
			return err
		}
	}
	return err
}

func (c File) CopyTo(writer io.Writer, offer *api.Offer) (err error) {
	for _, info := range offer.Files {
		if info.IsDir {
			continue
		}
		if err = c.copyFileTo(writer, offer, info); err != nil {
			return
		}
	}
	return
}

func (c File) copyFileTo(writer io.Writer, offer *api.Offer, info api.Info) (err error) {
	reader, err := c.resolve().Reader(info.Path)
	if err != nil {
		c.Println("Cannot get reader", info.Path, offer.Id, err)
		return
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 200 * time.Millisecond,
		DrawFunc: func(progress int64, size int64) error {
			status := fmt.Sprintf("upload %s %d/%dB", info.Path, progress, size)
			Outgoing(api.Core(c)).Update(offer, status, false)
			return nil
		},
	}
	_, err = io.CopyN(writer, progress, info.Size)
	if err != nil {
		c.Println("Cannot copy", info.Path, err)
		return
	}
	return
}

func (c File) resolve() api.FileResolver {
	if c.RemoteResolver {
		return remote.Resolver{}
	} else {
		return file.Resolver{}
	}
}
