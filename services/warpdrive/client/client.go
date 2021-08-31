package client

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"github.com/cryptopunkscc/astrald/services/warpdrive"
	"io"
	"log"
)

const tag = "warpdrive-client"

type WarpDriveClient struct {
	ctx  context.Context
	core api.Core
}

func NewWarpDriveClient(ctx context.Context, core api.Core) *WarpDriveClient {
	return &WarpDriveClient{ctx: ctx, core: core}
}

func (c *WarpDriveClient) ListPeers() ([]string, error) {
	s, err := connect.LocalRequest(c.ctx, c.core, warpdrive.PortLocal, warpdrive.List)
	if err != nil {
		return nil, err
	}

	size, err := s.ReadUint8()
	if err != nil {
		return nil, err
	}

	peers := make([]string, size)
	for i := uint8(0); i < size; i++ {
		peers[i], err = s.ReadStringWithSize8()
		if err != nil {
			return nil, err
		}
	}
	return peers, nil
}

func (c *WarpDriveClient) Send(identity api.Identity, filePath string) error {
	s, err := connect.LocalRequest(c.ctx, c.core, warpdrive.PortLocal, warpdrive.SendPath)
	if err != nil {
		return err
	}

	_, err = s.WriteStringWithSize8(string(identity))
	if err != nil {
		return err
	}

	_, err = s.WriteStringWithSize16(filePath)
	if err != nil {
		return err
	}

	_, err = s.ReadByte()
	if err != nil {
		return err
	}

	return nil
}

func (c *WarpDriveClient) Writer(nodeId api.Identity, fileName string) (io.WriteCloser, error) {
	s, err := connect.LocalRequest(c.ctx, c.core, warpdrive.PortLocal, warpdrive.SendStream)
	if err != nil {
		return nil, err
	}

	_, err = s.WriteStringWithSize8(string(nodeId))
	if err != nil {
		return nil, err
	}

	_, err = s.WriteStringWithSize16(fileName)
	if err != nil {
		return nil, err
	}

	res, err := s.ReadByte()
	if err != nil {
		log.Println(tag, "cannot read response", err)
		return nil, err
	}
	if res != warpdrive.Ok {
		log.Println(tag, "sending rejected by peer", res)
		return nil, errors.New("rejected")
	}

	return s, nil
}
