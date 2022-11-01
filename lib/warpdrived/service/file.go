package service

import (
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage/file"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"github.com/mitchellh/ioprogress"
	"io"
	"time"
)

var _ warpdrive.FileService = File{}

type File core.Component

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
	for i := range offer.Files {
		if i < offer.Index {
			continue
		}
		offer.Index = i
		if err = co.fileFrom(reader); err != nil {
			return
		}
		offer.Progress = 0
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
		co.update(offer, info.Size)
		return nil
	}
	offset := offer.Progress
	writer, err := s.FileWriter(info.Uri, info.Perm, offset)
	if err != nil {
		co.Println("Cannot get writer for", info.Uri, err)
		return
	}
	defer func() {
		err := writer.Close()
		if err != nil {
			co.Println("Cannot close info", info.Uri, err)
			return
		}
	}()
	// Copy bytes
	update := func(progress int64, size int64) error {
		co.update(offer, offset+progress)
		return nil
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 1000 * time.Millisecond,
		DrawFunc:     update,
	}
	l, err := io.CopyN(writer, progress, info.Size-offset)
	co.Progress = offset + l
	if err != nil {
		co.Println("Cannot read", info.Uri, err, "expected size", info.Size, "but was", co.Progress)
		return err
	}
	if co.Progress != info.Size {
		co.update(offer, info.Size)
	}
	return
}

func (co CopyOffer) To(writer io.Writer) (err error) {
	co.Status = warpdrive.StatusUpdated
	for i := range co.Files {
		if i < co.Index {
			continue
		}
		co.Index = i
		if err = co.fileTo(writer); err != nil {
			return
		}
		co.Progress = 0
	}
	return
}

func (co CopyOffer) fileTo(writer io.Writer) (err error) {
	info := co.Files[co.Index]
	if info.IsDir {
		co.update(co.Offer.Offer, info.Size)
		return
	}
	offset := co.Progress
	reader, err := co.Reader(info.Uri, offset)
	if err != nil {
		co.Println("Cannot get reader", info.Uri, co.Id, err)
		return
	}
	defer reader.Close()
	update := func(progress int64, size int64) error {
		co.update(co.Offer.Offer, offset+progress)
		return nil
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 1000 * time.Millisecond,
		DrawFunc:     update,
	}
	l, err := io.CopyN(writer, progress, info.Size-offset)
	if err != nil {
		co.Println("Cannot write", info.Uri, err)
		return err
	}
	co.Progress = offset + l
	if co.Progress != info.Size {
		co.update(co.Offer.Offer, info.Size)
	}
	return
}
