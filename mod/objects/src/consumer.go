package objects

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strconv"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Consumer = &Consumer{}

type Consumer struct {
	mod        *Module
	consumerID *astral.Identity
	providerID *astral.Identity
	*query.Point
}

func NewConsumer(mod *Module, consumerID *astral.Identity, providerID *astral.Identity) *Consumer {
	return &Consumer{
		mod:        mod,
		consumerID: consumerID,
		providerID: providerID,
		Point:      query.NewPoint(mod.node, consumerID, providerID),
	}
}

func (c *Consumer) Read(ctx *astral.Context, objectID *astral.ObjectID, offset int64, limit int64) (io.ReadCloser, error) {
	var q = query.New(ctx.Identity(), c.providerID, methodRead, &opReadArgs{
		ID:     objectID,
		Offset: astral.Uint64(offset),
		Limit:  astral.Uint64(limit),
	})

	conn, err := query.Route(ctx, c.mod.node, q)
	if err != nil {
		return nil, err
	}

	r := &NetworkReader{
		mod:        c.mod,
		objectID:   objectID,
		consumer:   c.consumerID,
		provider:   c.providerID,
		pos:        offset,
		ReadCloser: conn,
	}

	return r, nil
}

func (c *Consumer) Search(ctx *astral.Context, s string) (<-chan *objects.SearchResult, error) {
	params := query.Args{
		"q": s,
	}

	conn, err := query.Route(ctx,
		c.mod.node,
		query.New(c.consumerID, c.providerID, methodSearch, params),
	)
	if err != nil {
		return nil, err
	}
	ch := astral.NewChannel(conn)

	var results = make(chan *objects.SearchResult)

	go func() {
		<-ctx.Done()
		ch.Close()
	}()

	go func() {
		defer ch.Close()
		defer close(results)

		for {
			obj, err := ch.Read()
			if err != nil {
				return
			}
			switch obj := obj.(type) {
			case *objects.SearchResult:
				results <- obj

			default:
				c.mod.log.Errorv(2, "unexpected object type: %v", obj.ObjectType())
				return
			}
		}
	}()

	return results, nil
}

func (c *Consumer) Push(ctx *astral.Context, o astral.Object) (err error) {
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

	params := query.Args{
		"size": strconv.FormatInt(int64(len(b)), 10),
	}

	var q = query.New(c.consumerID, c.providerID, methodPush, params)

	conn, err := query.Route(ctx, c.mod.node, q)
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

func (c *Consumer) Describe(ctx *astral.Context, objectID *astral.ObjectID, _ *astral.Scope) (<-chan *objects.SourcedObject, error) {
	var results = make(chan *objects.SourcedObject, 1)

	var q = query.New(
		c.consumerID,
		c.providerID,
		methodDescribe,
		&describeArgs{ID: objectID})

	go func() {
		defer close(results)

		conn, err := query.Route(ctx, c.mod.node, q)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			obj, _, err := c.mod.Blueprints().ReadCanonical(conn)
			if err != nil {
				return
			}

			results <- &objects.SourcedObject{
				Source: c.providerID,
				Object: obj,
			}
		}
	}()

	return results, nil
}
