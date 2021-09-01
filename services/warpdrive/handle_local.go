package warpdrive

import (
	"errors"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"github.com/cryptopunkscc/astrald/services/warpdrive/peers"
	"io"
	"log"
	"path/filepath"
)

func (srv localService) listPeers(rc request.Context) error {

	p, err := peers.Get()
	if err != nil {
		return err
	}

	err = rc.WriteByte(byte(len(p)))
	if err != nil {
		return err
	}

	for _, id := range p {
		_, err = rc.WriteStringWithSize8(id)
		if err != nil {
			return err
		}
	}

	_, err = rc.ReadByte()
	if err != nil {
		return err
	}

	return nil
}

func (srv localService) sendFromStream(rc request.Context) error {

	// Read recipient id
	log.Println(rc.Port, "reading recipient id")
	rid, err := rc.ReadStringWithSize8()
	if err != nil {
		log.Println(rc.Port, "cannot read recipient id", err)
		return err
	}

	// Read file name
	log.Println(rc.Port, "reading file name")
	fn, err := rc.ReadStringWithSize16()
	if err != nil {
		log.Println(rc.Port, "cannot read file name", err)
		return err
	}

	// Connect to remote service
	log.Println(rc.Port, "connecting to", PortRemote)
	remote, err := connect.Remote(srv.ctx, srv.core, api.Identity(rid), PortRemote)
	if err != nil {
		log.Println(rc.Port, "cannot connect to", PortRemote, err)
		return err
	}

	// Send file name
	log.Println(rc.Port, "sending file name", fn, rid)
	_, err = remote.WriteStringWithSize16(fn)
	if err != nil {
		log.Println(rc.Port, "cannot send file name", rid, err)
		return err
	}

	// Reading response
	log.Println(rc.Port, "reading response", fn, rid)
	res, err := remote.ReadByte()
	if err != nil {
		log.Println(rc.Port, "cannot read response", err)
		return err
	}
	_ = rc.WriteByte(res)
	if res != Ok {
		log.Println(rc.Port, "sending rejected by peer", res)
		return errors.New("rejected")
	}

	// Send file
	log.Println(rc.Port, "sending file", fn, rid)
	_, err = io.Copy(remote, rc)
	if err != nil {
		log.Println(rc.Port, "cannot send file", fn, rid, err)
		return err
	}

	log.Println(rc.Port, "closing stream", fn, rid)
	err = remote.Close()
	if err != nil {
		log.Println(rc.Port, "closing stream", fn, rid, err)
		return err
	}

	return nil
}

func (srv localService) sendFromPath(rc request.Context) error {

	// Read recipient id
	log.Println(rc.Port, "reading recipient id")
	rid, err := rc.ReadStringWithSize8()
	if err != nil {
		log.Println(rc.Port, "cannot read recipient id", err)
		return err
	}

	// Read file path
	log.Println(rc.Port, "reading file path")
	fp, err := rc.ReadStringWithSize16()
	if err != nil {
		log.Println(rc.Port, "cannot read file path", err)
		return err
	}

	// Obtain file reader
	log.Println(rc.Port, "getting file reader")
	r, err := srv.store.Reader(fp)
	if err != nil {
		log.Println(rc.Port, "cannot get file reader", err)
		return err
	}

	// Get file name
	fn := filepath.Base(fp)

	// Connect to remote service
	log.Println(rc.Port, "connecting to", PortRemote)
	remote, err := connect.Remote(srv.ctx, srv.core, api.Identity(rid), PortRemote)
	if err != nil {
		log.Println(rc.Port, "cannot connect to", PortRemote)
		return err
	}

	// Send file name
	log.Println(rc.Port, "sending file name", fn, rid)
	_, err = remote.WriteStringWithSize16(fn)
	if err != nil {
		log.Println(rc.Port, "cannot send file name", rid, err)
		return err
	}

	// Reading response
	log.Println(rc.Port, "reading response", fn, rid)
	res, err := remote.ReadByte()
	if err != nil {
		log.Println(rc.Port, "cannot read response", err)
		return err
	}
	_ = rc.WriteByte(res)
	if res != Ok {
		log.Println(rc.Port, "sending rejected by peer", res)
		return errors.New("rejected")
	}

	// Send file
	log.Println(rc.Port, "sending file", fp, rid)
	_, err = io.Copy(remote, r)
	if err != nil {
		log.Println(rc.Port, "cannot send file", fp, rid, err)
		return err
	}

	log.Println(rc.Port, "closing stream", fp, rid)
	err = remote.Close()
	if err != nil {
		log.Println(rc.Port, "closing stream", fp, rid, err)
		return err
	}

	// Write ok response
	log.Println(rc.Port, "writing ok")
	err = rc.WriteByte(0)
	if err != nil {
		log.Println(rc.Port, "cannot write ok", err)
		return err
	}

	return nil
}
