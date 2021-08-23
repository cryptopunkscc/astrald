package sync

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/services/util/connect"
)

type Client struct {
	ctx context.Context
	core api.Core
}

func NewClient(ctx context.Context, core api.Core) *Client {
	return &Client{ctx: ctx, core: core}
}

func (c *Client) Download(nodeId api.Identity, fileId fid.ID) error  {
	stream, err := connect.LocalRequest(c.ctx, c.core, Port, Download)
	if err != nil {
		return fmt.Errorf("cannot request download %s %s %s", nodeId, fileId, err)
	}

	_, err = stream.WriteStringWithSize8(string(nodeId))
	if err != nil {
		return fmt.Errorf("cannot write node id %s %s", nodeId, err)
	}

	err = fileId.Write(stream)
	if err != nil {
		return fmt.Errorf("cannot write node id %s %s", fileId, err)
	}

	_, err = stream.ReadByte()
	if err != nil {
		return fmt.Errorf("download error %s %s %s", nodeId, fileId, err)
	}
	return nil
}