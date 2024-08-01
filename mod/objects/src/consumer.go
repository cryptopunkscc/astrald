package objects

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"strconv"
)

var _ objects.Consumer = &Consumer{}

type Consumer struct {
	mod        *Module
	consumerID id.Identity
	providerID id.Identity
}

func NewConsumer(mod *Module, consumerID id.Identity, providerID id.Identity) *Consumer {
	return &Consumer{
		mod:        mod,
		consumerID: consumerID,
		providerID: providerID,
	}
}

func (c *Consumer) Describe(ctx context.Context, objectID object.ID, _ *desc.Opts) (descs []*desc.Desc, err error) {
	var query = astral.NewQuery(
		c.consumerID,
		c.providerID,
		core.Query(
			methodDescribe,
			core.Params{
				"id": objectID.String(),
			},
		),
	)

	conn, err := astral.Route(ctx, c.mod.node, query)
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
			Source: c.providerID,
			Data:   d,
		})
	}

	return descs, nil
}

func (c *Consumer) Open(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (r objects.Reader, err error) {
	params := core.Params{
		"id": objectID.String(),
	}

	if opts.QueryFilter != nil {
		if !opts.QueryFilter(c.providerID) {
			return
		}
	}

	if opts.Offset != 0 {
		params.SetInt("offset", int(opts.Offset))
	}

	var query = astral.NewQuery(c.consumerID, c.providerID, core.Query(methodRead, params))

	conn, err := astral.Route(ctx, c.mod.node, query)
	if err != nil {
		return nil, err
	}

	r = &NetworkReader{
		mod:        c.mod,
		objectID:   objectID,
		consumer:   c.consumerID,
		provider:   c.providerID,
		pos:        int64(opts.Offset),
		ReadCloser: conn,
	}

	return
}

func (c *Consumer) Put(ctx context.Context, p []byte) (object.ID, error) {
	params := core.Params{
		"size": strconv.FormatInt(int64(len(p)), 10),
	}

	var query = astral.NewQuery(c.consumerID, c.providerID, core.Query(methodPut, params))

	conn, err := astral.Route(ctx, c.mod.node, query)
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
	params := core.Params{
		"q": q,
	}

	var query = astral.NewQuery(c.consumerID, c.providerID, core.Query(methodSearch, params))

	conn, err := astral.Route(ctx, c.mod.node, query)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = json.NewDecoder(conn).Decode(&matches)

	return
}

func (c *Consumer) Push(ctx context.Context, o astral.Object) (err error) {
	var buf = &bytes.Buffer{}

	_, err = astral.ObjectHeader(o.ObjectType()).WriteTo(buf)
	if err != nil {
		return
	}

	_, err = o.WriteTo(buf)
	if err != nil {
		return
	}

	var b = buf.Bytes()
	if len(b) > maxPushSize {
		return errors.New("object too large")
	}

	params := core.Params{
		"size": strconv.FormatInt(int64(len(b)), 10),
	}

	var query = astral.NewQuery(c.consumerID, c.providerID, core.Query(methodPush, params))

	conn, err := astral.Route(ctx, c.mod.node, query)
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = conn.Write(b)
	if err != nil {
		return
	}

	var ok bool
	err = binary.Read(conn, binary.BigEndian, &ok)
	if err != nil {
		return
	}
	if !ok {
		return errors.New("object rejected")
	}

	return nil
}
