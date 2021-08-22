package share

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/shares"
	"github.com/cryptopunkscc/astrald/components/sio"
	"log"
)

type Context struct {
	shares shares.Shares
}

func NewContext(shares shares.Shares) *Context {
	return &Context{shares: shares}
}

func (r *Context) Add(
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

func (r *Context) Remove(
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

func (r *Context) List(
	_ api.Identity,
	query string,
	stream sio.ReadWriteCloser,
) error {
	id, err := stream.ReadStringWithSize8()
	if err != nil {
		log.Println(query, "cannot read id", err)
		return err
	}
	s, err := r.shares.List(api.Identity(id))
	if err != nil {
		log.Println(query, "cannot list shares for", id, err)
		return err
	}
	_, err = stream.WriteUInt32(uint32(len(s)))
	if err != nil {
		log.Println(query, "cannot send shares count", id, err)
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

func (r *Context) Contains(
	_ api.Identity,
	query string,
	stream sio.ReadWriteCloser,
) error {
	log.Println(query, "reading identity")
	nodeId, err := stream.ReadStringWithSize8()
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
	log.Println(query, "checking if contains", nodeId, fileId)
	contains, err := r.shares.Contains(api.Identity(nodeId), fileId)
	if err != nil {
		log.Println(query, "cannot check contains share", nodeId, fileId)
		return err
	}
	response := byte(0)
	if contains {
		response = 1
	}
	log.Println(query, "sending response", response, nodeId, fileId)
	err = stream.WriteByte(response)
	if err != nil {
		log.Println(query, "cannot send response", response, nodeId, fileId)
		return err
	}
	log.Println(query, "done contains", response, nodeId, fileId)
	return nil
}
