package service

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/file"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/remote"
	"github.com/mitchellh/ioprogress"
	"io"
	"time"
)

type File api.Core
type CopyOffer struct {
	File
	*Offer
}

func (f File) Info(uri string) (files []api.Info, err error) {
	return f.resolve().Info(uri)
}

func (f File) Copy(offer *Offer) CopyOffer {
	return CopyOffer{File: f, Offer: offer}
}

func (offer CopyOffer) From(reader io.Reader) (err error) {
	offer.Status = api.StatusUpdated
	for offer.Index = range offer.Files {
		if err = offer.fileFrom(reader); err != nil {
			return
		}
	}
	return
}

func (offer CopyOffer) fileFrom(reader io.Reader) (err error) {
	storage := file.Storage(offer.File)
	info := offer.Files[offer.Index]
	if info.IsDir {
		err := storage.MkDir(info.Uri, info.Perm)
		if err != nil && !storage.IsExist(err) {
			offer.Println("Cannot make dir", info.Uri, err)
			return err
		}
		offer.Update(info.Size)
		return nil
	}
	writer, err := storage.FileWriter(info.Uri, info.Perm)
	if err != nil {
		offer.Println("Cannot get writer for", info.Uri, err)
		return
	}
	defer writer.Close()
	// Copy bytes
	update := func(progress int64, size int64) error {
		offer.Update(progress)
		return nil
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 1000 * time.Millisecond,
		DrawFunc:     update,
	}
	l, err := io.CopyN(writer, progress, info.Size)
	if err != nil {
		offer.Println("Cannot read", info.Uri, err, "expected size", info.Size, "but was", l)
		return
	}
	if offer.Progress != info.Size {
		offer.Update(info.Size)
	}
	err = writer.Close()
	if err != nil {
		offer.Println("Cannot close info", info.Uri, err)
		return
	}
	return
}

func (offer CopyOffer) To(writer io.Writer) (err error) {
	offer.Status = api.StatusUpdated
	for offer.Index = range offer.Files {
		if err = offer.fileTo(writer); err != nil {
			return
		}
	}
	return
}

func (offer CopyOffer) fileTo(writer io.Writer) (err error) {
	info := offer.Files[offer.Index]
	if info.IsDir {
		offer.Update(info.Size)
		return
	}
	reader, err := offer.resolve().Reader(info.Uri)
	if err != nil {
		offer.Println("Cannot get reader", info.Uri, offer.Id, err)
		return
	}
	update := func(progress int64, size int64) error {
		offer.Update(progress)
		return nil
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 1000 * time.Millisecond,
		DrawFunc:     update,
	}
	_, err = io.CopyN(writer, progress, info.Size)
	if err != nil {
		offer.Println("Cannot write", info.Uri, err)
		return
	}
	if offer.Progress != info.Size {
		offer.Update(info.Size)
	}
	return
}

func (f File) resolve() api.FileResolver {
	if f.RemoteResolver {
		return remote.Resolver{}
	} else {
		return file.Resolver{}
	}
}
