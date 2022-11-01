package service

import (
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage/file"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"github.com/mitchellh/ioprogress"
	"io"
	"time"
)

func (srv *offer) Copy(offer *warpdrive.Offer) warpdrive.CopyOffer {
	srv.Offer = offer
	return srv
}

func (srv *offer) From(reader io.Reader) (err error) {
	offer := srv.Offer
	offer.Status = warpdrive.StatusUpdated
	for i := range offer.Files {
		if i < offer.Index {
			continue
		}
		offer.Index = i
		if err = srv.fileFrom(reader); err != nil {
			return
		}
		offer.Progress = 0
	}
	return
}

func (srv *offer) fileFrom(reader io.Reader) (err error) {
	s := file.Storage(srv.Component)
	offer := srv.Offer

	info := srv.Files[offer.Index]
	if info.IsDir {
		err := s.MkDir(info.Uri, info.Perm)
		if err != nil && !s.IsExist(err) {
			srv.Println("Cannot make dir", info.Uri, err)
			return err
		}
		srv.update(info.Size)
		return nil
	}
	offset := offer.Progress
	writer, err := s.FileWriter(info.Uri, info.Perm, offset)
	if err != nil {
		srv.Println("Cannot get writer for", info.Uri, err)
		return
	}
	defer func() {
		err := writer.Close()
		if err != nil {
			srv.Println("Cannot close info", info.Uri, err)
			return
		}
	}()
	// Copy bytes
	update := func(progress int64, size int64) error {
		srv.update(offset + progress)
		return nil
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 1000 * time.Millisecond,
		DrawFunc:     update,
	}
	l, err := io.CopyN(writer, progress, info.Size-offset)
	srv.Progress = offset + l
	if err != nil {
		srv.Println("Cannot read", info.Uri, err, "expected size", info.Size, "but was", srv.Progress)
		return err
	}
	if srv.Progress != info.Size {
		srv.update(info.Size)
	}
	return
}

func (srv *offer) To(writer io.Writer) (err error) {
	srv.Status = warpdrive.StatusUpdated
	for i := range srv.Files {
		if i < srv.Index {
			continue
		}
		srv.Index = i
		if err = srv.fileTo(writer); err != nil {
			return
		}
		srv.Progress = 0
	}
	return
}

func (srv *offer) fileTo(writer io.Writer) (err error) {
	info := srv.Files[srv.Index]
	if info.IsDir {
		srv.update(info.Size)
		return
	}
	offset := srv.Progress
	reader, err := srv.Reader(info.Uri, offset)
	if err != nil {
		srv.Println("Cannot get reader", info.Uri, srv.Id, err)
		return
	}
	defer reader.Close()
	update := func(progress int64, size int64) error {
		srv.update(offset + progress)
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
		srv.Println("Cannot write", info.Uri, err)
		return err
	}
	srv.Progress = offset + l
	if srv.Progress != info.Size {
		srv.update(info.Size)
	}
	return
}

func (srv *offer) update(progress int64) {
	srv.Progress = progress
	srv.dispatch(srv.Offer)
}
