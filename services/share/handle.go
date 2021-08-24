package share

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

func (srv *service) Add(rc request.Context) error {
	// Read identity
	log.Println(rc.Port, "reading identity")
	nodeId, err := rc.ReadWithSize8()
	if err != nil {
		log.Println(rc.Port, "cannot read identity", err)
		return err
	}

	// Read file id
	log.Println(rc.Port, "reading file id")
	fileId, _, err := fid.Read(rc)
	if err != nil {
		log.Println(rc.Port, "cannot read file id", err)
		return err
	}

	// Add share
	log.Println(rc.Port, "adding share", nodeId, fileId)
	err = srv.shared.Add(api.Identity(nodeId), fileId)
	if err != nil {
		log.Println(rc.Port, "cannot add share", nodeId, fileId)
		return err
	}

	// Send ok
	log.Println(rc.Port, "sending ok")
	err = rc.WriteByte(0)
	if err != nil {
		log.Println(rc.Port, "cannot send ok", err)
		return err
	}
	log.Println(rc.Port, "finish adding share")
	return nil
}

func (srv *service) Remove(rc request.Context) error {
	log.Println(rc.Port, "reading identity")
	nodeId, err := rc.ReadWithSize8()
	if err != nil {
		log.Println(rc.Port, "cannot read identity", err)
		return err
	}
	log.Println(rc.Port, "reading file id")
	fileId, _, err := fid.Read(rc)
	if err != nil {
		log.Println(rc.Port, "cannot read file id", err)
		return err
	}
	log.Println(rc.Port, "removing share", nodeId, fileId)
	err = srv.shared.Remove(api.Identity(nodeId), fileId)
	if err != nil {
		log.Println(rc.Port, "cannot remove share", nodeId, fileId)
		return err
	}
	return nil
}

func (srv *service) ListLocal(rc request.Context) error {
	id, err := rc.ReadStringWithSize8()
	if err != nil {
		log.Println(rc.Port, "cannot read id", err)
		return err
	}
	rc.Caller = api.Identity(id)
	return srv.List(rc)
}

func (srv *service) List(rc request.Context) error {
	list, err := srv.shared.List(rc.Caller)
	if err != nil {
		log.Println(rc.Port, "cannot list shares for", rc.Caller, err)
		return err
	}
	_, err = rc.WriteUInt32(uint32(len(list)))
	if err != nil {
		log.Println(rc.Port, "cannot send shares count", rc.Caller, err)
		return err
	}
	for _, share := range list {
		err = share.Write(rc)
		if err != nil {
			log.Println(rc.Port, "cannot send id", err)
			return err
		}
	}
	return nil
}

func (srv *service) ContainsLocal(rc request.Context) error {
	id, err := rc.ReadStringWithSize8()
	if err != nil {
		log.Println(rc.Port, "cannot read id", err)
		return err
	}
	rc.Caller = api.Identity(id)
	return srv.List(rc)
}

func (srv *service) Contains(rc request.Context) error {
	log.Println(rc.Port, "reading file id")
	fileId, _, err := fid.Read(rc)
	if err != nil {
		log.Println(rc.Port, "cannot read file id", err)
		return err
	}
	log.Println(rc.Port, "checking if contains", rc.Caller, fileId)
	contains, err := srv.shared.Contains(rc.Caller, fileId)
	if err != nil {
		log.Println(rc.Port, "cannot check contains share", rc.Caller, fileId)
		return err
	}
	response := byte(0)
	if contains {
		response = 1
	}
	log.Println(rc.Port, "sending response", response, rc.Caller, fileId)
	err = rc.WriteByte(response)
	if err != nil {
		log.Println(rc.Port, "cannot send response", response, rc.Caller, fileId)
		return err
	}
	log.Println(rc.Port, "done contains", response, rc.Caller, fileId)
	return nil
}
