package share

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/shares"
	"github.com/cryptopunkscc/astrald/services/util/connect"
)

type client struct {
	context.Context
	api.Core
}

func NewClient(
	ctx context.Context,
	core api.Core,
) shares.Shares {
	return &client{ctx, core}
}

func (c *client) Add(id api.Identity, file fid.ID) error {
	stream, err := connect.LocalRequest(c, c, Port, Add)
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

func (c *client) Remove(id api.Identity, file fid.ID) error {
	stream, err := connect.LocalRequest(c, c, Port, Remove)
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

func (c *client) List(id api.Identity) ([]fid.ID, error) {
	stream, err := connect.LocalRequest(c, c, Port, List)
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

func (c *client) Contains(id api.Identity, file fid.ID) (bool, error) {
	stream, err := connect.LocalRequest(c, c, Port, Contains)
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
