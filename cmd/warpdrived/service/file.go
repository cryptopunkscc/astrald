package service

import (
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/core"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/storage"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/storage/file"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/storage/remote"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"github.com/mitchellh/ioprogress"
	"io"
	"time"
)

var _ warpdrive.FileService = File{}

type File core.Component

func (f File) Info(uri string) (files []warpdrive.Info, err error) {
	return f.resolve().Info(uri)
}

func (f File) resolve() storage.FileResolver {
	if f.RemoteResolver {
		return remote.Resolver{}
	} else {
		return file.Resolver{}
	}
}

func (f File) Copy(offer *warpdrive.Offer) warpdrive.CopyOffer {
	cp := CopyOffer{File: f}
	if offer.In {
		cp.Offer = Incoming(core.Component(f))
	} else {
		cp.Offer = Outgoing(core.Component(f))
	}
	cp.Offer.Offer = offer
	return cp
}

type CopyOffer struct {
	File
	*Offer
}

func (co CopyOffer) From(reader io.Reader) (err error) {
	offer := co.Offer.Offer
	offer.Status = warpdrive.StatusUpdated
	for offer.Index = range offer.Files {
		if err = co.fileFrom(reader); err != nil {
			return
		}
	}
	return
}

func (co CopyOffer) fileFrom(reader io.Reader) (err error) {
	s := file.Storage(co.File)
	offer := co.Offer.Offer

	info := offer.Files[offer.Index]
	if info.IsDir {
		err := s.MkDir(info.Uri, info.Perm)
		if err != nil && !s.IsExist(err) {
			co.Println("Cannot make dir", info.Uri, err)
			return err
		}
		co.Update(offer, info.Size)
		return nil
	}
	writer, err := s.FileWriter(info.Uri, info.Perm)
	if err != nil {
		co.Println("Cannot get writer for", info.Uri, err)
		return
	}
	defer writer.Close()
	// Copy bytes
	update := func(progress int64, size int64) error {
		co.Update(offer, progress)
		return nil
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 1000 * time.Millisecond,
		DrawFunc:     update,
	}
	co.Progress, err = io.CopyN(writer, progress, info.Size)
	if err != nil {
		co.Println("Cannot read", info.Uri, err, "expected size", info.Size, "but was", co.Progress)
		return
	}
	if co.Progress != info.Size {
		co.Update(offer, info.Size)
	}
	err = writer.Close()
	if err != nil {
		co.Println("Cannot close info", info.Uri, err)
		return
	}
	return
}

func (co CopyOffer) To(writer io.Writer) (err error) {
	co.Status = warpdrive.StatusUpdated
	for co.Index = range co.Files {
		if err = co.fileTo(writer); err != nil {
			return
		}
	}
	return
}

func (co CopyOffer) fileTo(writer io.Writer) (err error) {
	info := co.Files[co.Index]
	if info.IsDir {
		co.Update(co.Offer.Offer, info.Size)
		return
	}
	reader, err := co.resolve().Reader(info.Uri)
	if err != nil {
		co.Println("Cannot get reader", info.Uri, co.Id, err)
		return
	}
	defer reader.Close()
	update := func(progress int64, size int64) error {
		co.Update(co.Offer.Offer, progress)
		return nil
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 1000 * time.Millisecond,
		DrawFunc:     update,
	}
	co.Progress, err = io.CopyN(writer, progress, info.Size)
	if err != nil {
		co.Println("Cannot write", info.Uri, err)
		return
	}
	if co.Progress != info.Size {
		co.Update(co.Offer.Offer, info.Size)
	}
	return
}
