package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/proto/block"
	_store "github.com/cryptopunkscc/astrald/proto/store"
	"io"
)

type Client struct {
	nodeID   id.Identity
	nodeName string
}

func NewClient(nodeName string) *Client {
	return &Client{
		nodeName: nodeName,
	}
}

var localnode = NewClient("localnode")

func (client *Client) Open(blockID string, flags uint32) (block.Block, error) {
	parsedID, err := data.Parse(blockID)
	if err != nil {
		return nil, err
	}

	conn, err := astral.QueryName(client.nodeName, "storage")
	if err != nil {
		panic(err)
	}

	store := _store.Bind(conn)

	return store.Open(parsedID, flags)
}

func (client *Client) Download(blockID string, offset uint64, limit uint64) (io.ReadCloser, error) {
	parsedID, err := data.Parse(blockID)
	if err != nil {
		return nil, err
	}

	conn, err := astral.QueryName(client.nodeName, "storage")
	if err != nil {
		return nil, err
	}

	store := _store.Bind(conn)

	return store.Download(parsedID, offset, limit)
}

func Open(blockID string, flags uint32) (block.Block, error) {
	return localnode.Open(blockID, flags)
}

func Download(blockID string, offset uint64, limit uint64) (io.ReadCloser, error) {
	return localnode.Download(blockID, offset, limit)
}
