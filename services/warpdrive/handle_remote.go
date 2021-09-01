package warpdrive

import (
	"github.com/cryptopunkscc/astrald/services/util/request"
	"io"
	"log"
)

func (srv remoteService) receive(rc request.Context) error {

	// Read file name
	log.Println(rc.Port, "reading file name")
	fn, err := rc.ReadStringWithSize16()
	if err != nil {
		log.Println(rc.Port, "cannot read file name", err)
		return err
	}

	// Check if file exist
	log.Println(rc.Port, "checking if file exist", fn)
	_, err = srv.store.Reader(fn)
	if err == nil {
		log.Println(rc.Port, "file already exist", fn)
		_ = rc.WriteByte(Rejected)
		return nil
	}

	// Send ok
	log.Println(rc.Port, "sending ok", fn)
	err = rc.WriteByte(Ok)
	if err != nil {
		log.Println(rc.Port, "cannot send ok", fn, err)
		return err
	}

	// Get local writer
	log.Println(rc.Port, "getting writer", fn)
	w, err := srv.store.Writer()
	if err != nil {
		log.Println(rc.Port, "cannot get writer", fn, err)
		return err
	}

	// Copy file
	log.Println(rc.Port, "saving file", fn)
	_, err = io.Copy(w, rc)
	if err != nil {
		log.Println(rc.Port, "cannot save file", fn)
		return err
	}

	// Rename file
	log.Println(rc.Port, "renaming file", fn)
	err = w.Rename(fn)
	if err != nil {
		log.Println(rc.Port, "cannot rename file", fn)
		return err
	}

	return nil
}
