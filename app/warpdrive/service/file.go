package service

import (
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
	for index := range offer.Files {
		if err = c.copyFileFrom(reader, offer, index); err != nil {
			return
		}
	}
	return
}

func (c File) copyFileFrom(reader io.Reader, offer *api.Offer, index int) (err error) {
	info := offer.Files[index]
	storage := file.Storage(c)
	if info.IsDir {
		err := storage.MkDir(info.Uri, info.Perm)
		if err != nil && !storage.IsExist(err) {
			c.Println("Cannot make dir", info.Uri, err)
			return err
		}
	} else {
		writer, err := storage.FileWriter(info.Uri, info.Perm)
		if err != nil {
			c.Println("Cannot get writer for", info.Uri, err)
			return err
		}
		defer writer.Close()
		// Copy bytes
		offer.Status.Status = api.StatusProgress
		incoming := Incoming(api.Core(c))
		update := func(progress int64, size int64) error {
			info.Progress = progress
			offer.Files[index] = info
			incoming.Update(offer, index)
			return nil
		}
		progress := &ioprogress.Reader{
			Reader:       reader,
			Size:         info.Size,
			DrawInterval: 200 * time.Millisecond,
			DrawFunc:     update,
		}
		_, err = io.CopyN(writer, progress, info.Size)
		if err != nil {
			c.Println("Cannot copy", info.Uri, err)
			return err
		}
		err = writer.Close()
		if err != nil {
			c.Println("Cannot close info", info.Uri, err)
			return err
		}
	}
	return err
}

func (c File) CopyTo(writer io.Writer, offer *api.Offer) (err error) {
	for index, info := range offer.Files {
		if info.IsDir {
			continue
		}
		if err = c.copyFileTo(writer, offer, index); err != nil {
			return
		}
	}
	return
}

func (c File) copyFileTo(writer io.Writer, offer *api.Offer, index int) (err error) {
	info := offer.Files[index]
	reader, err := c.resolve().Reader(info.Uri)
	if err != nil {
		c.Println("Cannot get reader", info.Uri, offer.Id, err)
		return
	}
	offer.Status.Status = api.StatusProgress
	outgoing := Outgoing(api.Core(c))
	update := func(progress int64, size int64) error {
		info.Progress = progress
		offer.Files[index] = info
		outgoing.Update(offer, index)
		return nil
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 200 * time.Millisecond,
		DrawFunc:     update,
	}
	_, err = io.CopyN(writer, progress, info.Size)
	if err != nil {
		c.Println("Cannot copy", info.Uri, err)
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
