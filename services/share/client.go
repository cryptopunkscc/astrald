package share

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/shares"
	"github.com/cryptopunkscc/astrald/services/util/connect"
)

type sharedFilesClient struct {
	context.Context
	api.Core
	id api.Identity
	port string
}

func NewSharedFilesClient(
	ctx context.Context,
	core api.Core,
) shares.Shared {
	return &sharedFilesClient{ctx, core, "", Port}
}

func (c *sharedFilesClient) Add(id api.Identity, file fid.ID) error {
	stream, err := connect.RemoteRequest(c, c, c.id, c.port, Add)
	if err != nil {
		return err
	}
	_, err = stream.WriteStringWithSize8(string(id))
	if err != nil {
		return err
	}
	err = file.Write(stream)
	if err != nil {
		return err
	}
	_, err = stream.ReadByte()
	if err != nil {
		return err
	}
	return err
}

func (c *sharedFilesClient) Remove(id api.Identity, file fid.ID) error {
	stream, err := connect.RemoteRequest(c, c, c.id, c.port, Remove)
	if err != nil {
		return err
	}
	_, err = stream.WriteStringWithSize8(string(id))
	if err != nil {
		return err
	}
	err = file.Write(stream)
	if err != nil {
		return err
	}
	_, err = stream.ReadByte()
	if err != nil {
		return err
	}
	return nil
}

func (c *sharedFilesClient) List(id api.Identity) ([]fid.ID, error) {
	stream, err := connect.RemoteRequest(c, c, c.id, c.port, List)
	if err != nil {
		return nil, err
	}
	_, err = stream.WriteStringWithSize8(string(id))
	if err != nil {
		return nil, err
	}
	count, err := stream.ReadUint32()
	if err != nil {
		return nil, err
	}
	var buff [fid.Size]byte
	var files []fid.ID
	for i := uint32(0); i < count; i++ {
		_, err := stream.Read(buff[:])
		if err != nil {
			return nil, err
		}
		files = append(files, fid.Unpack(buff))
	}
	return files, nil
}

func (c *sharedFilesClient) Contains(id api.Identity, file fid.ID) (bool, error) {
	stream, err := connect.RemoteRequest(c, c, c.id, c.port, Contains)
	if err != nil {
		return false, err
	}
	_, err = stream.WriteStringWithSize8(string(id))
	if err != nil {
		return false, err
	}
	err = file.Write(stream)
	if err != nil {
		return false, err
	}
	contains, err := stream.ReadByte()
	if err != nil {
		return false, err
	}
	return contains == 1, nil
}

type sharesClient struct {
	context.Context
	api.Core
	id api.Identity
	port string
}

func NewSharesClient(
	ctx context.Context,
	core api.Core,
	id api.Identity,
) shares.Shares {
	return &sharesClient{ctx, core, id, RemotePort}
}

func (c *sharesClient) List() ([]fid.ID, error) {
	stream, err := connect.RemoteRequest(c, c, c.id, c.port, List)
	if err != nil {
		return nil, err
	}
	count, err := stream.ReadUint32()
	if err != nil {
		return nil, err
	}
	var buff [fid.Size]byte
	var files []fid.ID
	for i := uint32(0); i < count; i++ {
		_, err := stream.Read(buff[:])
		if err != nil {
			return nil, err
		}
		files = append(files, fid.Unpack(buff))
	}
	return files, nil
}

func (c *sharesClient) Contains(file fid.ID) (bool, error) {
	stream, err := connect.RemoteRequest(c, c, c.id, c.port, Contains)
	if err != nil {
		return false, err
	}
	err = file.Write(stream)
	if err != nil {
		return false, err
	}
	contains, err := stream.ReadByte()
	if err != nil {
		return false, err
	}
	return contains == 1, nil
}
