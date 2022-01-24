package core

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/mitchellh/ioprogress"
	"io"
	"time"
)

type fileManager core

func (c *core) File() api.FileManager {
	return (*fileManager)(c)
}

func (c *fileManager) CopyFrom(reader io.Reader, offer *api.Offer) (err error) {
	for _, file := range offer.Files {
		if err = c.copyFileFrom(reader, offer, file); err != nil {
			return
		}
	}
	return
}

func (c *fileManager) copyFileFrom(reader io.Reader, offer *api.Offer, file api.Info) (err error) {
	// Obtain writer
	if file.IsDir {
		err := c.MkDir(file.Path, file.Perm)
		if err != nil && !c.IsExist(err) {
			c.Println("Cannot make dir", file.Path, err)
			return err
		}
	} else {
		writer, err := c.FileWriter(file.Path, file.Perm)
		if err != nil {
			c.Println("Cannot get writer for", file.Path, err)
			return err
		}
		defer writer.Close()
		// Copy bytes
		progress := &ioprogress.Reader{
			Reader:       reader,
			Size:         file.Size,
			DrawInterval: 200 * time.Millisecond,
			DrawFunc: func(progress int64, size int64) error {
				status := fmt.Sprintf("download: %s %d/%dB", file.Path, progress, size)
				(*core)(c).Incoming().Update(offer, status, false)
				return nil
			},
		}
		_, err = io.CopyN(writer, progress, file.Size)
		if err != nil {
			c.Println("Cannot copy", file.Path, err)
			return err
		}
		err = writer.Close()
		if err != nil {
			c.Println("Cannot close file", file.Path, err)
			return err
		}
	}
	return err
}

func (c *fileManager) CopyTo(writer io.Writer, offer *api.Offer) (err error) {
	for _, file := range offer.Files {
		if file.IsDir {
			continue
		}
		if err = c.copyFileTo(writer, offer, file); err != nil {
			return
		}
	}
	return
}

func (c *fileManager) copyFileTo(writer io.Writer, offer *api.Offer, file api.Info) (err error) {
	reader, err := c.File(file.Path)
	if err != nil {
		c.Println("Cannot get reader", file.Path, offer.Id, err)
		return
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         file.Size,
		DrawInterval: 200 * time.Millisecond,
		DrawFunc: func(progress int64, size int64) error {
			status := fmt.Sprintf("upload %s %d/%dB", file.Path, progress, size)
			(*core)(c).Outgoing().Update(offer, status, false)
			return nil
		},
	}
	_, err = io.CopyN(writer, progress, file.Size)
	if err != nil {
		c.Println("Cannot copy", file.Path, err)
		return
	}
	return
}
