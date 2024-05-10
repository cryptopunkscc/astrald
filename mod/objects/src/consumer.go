package objects

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"strconv"
)

var _ objects.Consumer = &Consumer{}

type Consumer struct {
	mod    *Module
	caller id.Identity
	target id.Identity
}

func NewConsumer(mod *Module, caller id.Identity, target id.Identity) *Consumer {
	return &Consumer{mod: mod, caller: caller, target: target}
}

func (c *Consumer) Describe(ctx context.Context, objectID object.ID, _ *desc.Opts) (descs []*desc.Desc, err error) {
	var query = net.NewQuery(
		c.caller,
		c.target,
		router.Query(
			describeServiceName,
			router.Params{
				"id": objectID.String(),
			},
		),
	)

	conn, err := net.Route(ctx, c.mod.node.Router(), query)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	jbytes, err := io.ReadAll(conn)
	if err != nil {
		return nil, err
	}

	var j []JSONDescriptor
	err = json.Unmarshal(jbytes, &j)
	if err != nil {
		return nil, err
	}

	for _, i := range j {
		var d = c.mod.UnmarshalDescriptor(i.Type, i.Data)
		if d == nil {
			continue
		}

		descs = append(descs, &desc.Desc{
			Source: c.target,
			Data:   d,
		})
	}

	return descs, nil
}

func (c *Consumer) Open(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (conn net.SecureConn, err error) {
	params := router.Params{
		"id": objectID.String(),
	}

	if opts.QueryFilter != nil {
		if !opts.QueryFilter(c.target) {
			return
		}
	}

	if opts.Offset != 0 {
		params.SetInt("offset", int(opts.Offset))
	}

	var query = net.NewQuery(c.caller, c.target, router.Query(readServiceName, params))

	return net.Route(ctx, c.mod.node.Router(), query)
}

func (c *Consumer) Put(ctx context.Context, p []byte) (object.ID, error) {
	params := router.Params{
		"size": strconv.FormatInt(int64(len(p)), 10),
	}

	var query = net.NewQuery(c.caller, c.target, router.Query(putServiceName, params))

	conn, err := net.Route(ctx, c.mod.node.Router(), query)
	if err != nil {
		return object.ID{}, err
	}
	defer conn.Close()

	n, err := conn.Write(p)
	if err != nil {
		return object.ID{}, err
	}
	if n != len(p) {
		return object.ID{}, errors.New("write failed")
	}

	var status int
	var objectID object.ID

	err = cslq.Decode(conn, "cv", &status, &objectID)
	if err != nil {
		return object.ID{}, err
	}
	if status != 0 {
		return object.ID{}, errors.New("remote error")
	}

	return objectID, nil
}

func (c *Consumer) Search(ctx context.Context, q string) (matches []objects.Match, err error) {
	params := router.Params{
		"q": q,
	}

	var query = net.NewQuery(c.caller, c.target, router.Query(searchServiceName, params))

	conn, err := net.Route(ctx, c.mod.node.Router(), query)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = json.NewDecoder(conn).Decode(&matches)

	return
}
