package share

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/sio"
	"log"
)

func (r *requestContext) Add(
	_ api.Identity,
	query string,
	stream sio.ReadWriteCloser,
) error {
	// Read identity
	log.Println(query, "reading identity")
	nodeId, err := stream.ReadWithSize8()
	if err != nil {
		log.Println(query, "cannot read identity", err)
		return err
	}

	// Read file id
	log.Println(query, "reading file id")
	fileId, _, err := fid.Read(stream)
	if err != nil {
		log.Println(query, "cannot read file id", err)
		return err
	}

	// Add share
	log.Println(query, "adding share", nodeId, fileId)
	err = r.shares.Add(api.Identity(nodeId), fileId)
	if err != nil {
		log.Println(query, "cannot add share", nodeId, fileId)
		return err
	}

	// Send ok
	log.Println(query, "sending ok")
	err = stream.WriteByte(0)
	if err != nil {
		log.Println(query, "cannot send ok", err)
		return err
	}
	log.Println(query, "finish adding share")
	return nil
}

func (r *requestContext) Remove(
	_ api.Identity,
	query string,
	stream sio.ReadWriteCloser,
) error {
	log.Println(query, "reading identity")
	nodeId, err := stream.ReadWithSize8()
	if err != nil {
		log.Println(query, "cannot read identity", err)
		return err
	}
	log.Println(query, "reading file id")
	fileId, _, err := fid.Read(stream)
	if err != nil {
		log.Println(query, "cannot read file id", err)
		return err
	}
	log.Println(query, "removing share", nodeId, fileId)
	err = r.shares.Remove(api.Identity(nodeId), fileId)
	if err != nil {
		log.Println(query, "cannot remove share", nodeId, fileId)
		return err
	}
	return nil
}

func (r *requestContext) ListLocal(
	_ api.Identity,
	query string,
	stream sio.ReadWriteCloser,
) error {
	id, err := stream.ReadStringWithSize8()
	if err != nil {
		log.Println(query, "cannot read id", err)
		return err
	}
	return r.List(api.Identity(id), query, stream)
}

func (r *requestContext) List(
	caller api.Identity,
	query string,
	stream sio.ReadWriteCloser,
) error {
	s, err := r.shares.List(caller)
	if err != nil {
		log.Println(query, "cannot list shares for", caller, err)
		return err
	}
	_, err = stream.WriteUInt32(uint32(len(s)))
	if err != nil {
		log.Println(query, "cannot send shares count", caller, err)
		return err
	}
	for _, share := range s {
		err = share.Write(stream)
		if err != nil {
			log.Println(query, "cannot send id", err)
			return err
		}
	}
	return nil
}

func (r *requestContext) ContainsLocal(
	_ api.Identity,
	query string,
	stream sio.ReadWriteCloser,
) error {
	id, err := stream.ReadStringWithSize8()
	if err != nil {
		log.Println(query, "cannot read id", err)
		return err
	}
	return r.List(api.Identity(id), query, stream)
}

func (r *requestContext) Contains(
	caller api.Identity,
	query string,
	stream sio.ReadWriteCloser,
) error {
	log.Println(query, "reading file id")
	fileId, _, err := fid.Read(stream)
	if err != nil {
		log.Println(query, "cannot read file id", err)
		return err
	}
	log.Println(query, "checking if contains", caller, fileId)
	contains, err := r.shares.Contains(api.Identity(caller), fileId)
	if err != nil {
		log.Println(query, "cannot check contains share", caller, fileId)
		return err
	}
	response := byte(0)
	if contains {
		response = 1
	}
	log.Println(query, "sending response", response, caller, fileId)
	err = stream.WriteByte(response)
	if err != nil {
		log.Println(query, "cannot send response", response, caller, fileId)
		return err
	}
	log.Println(query, "done contains", response, caller, fileId)
	return nil
}
