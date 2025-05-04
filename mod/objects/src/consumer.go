package objects

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"strconv"
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

func (c *Consumer) Read(ctx *astral.Context, objectID *object.ID, offset int64, limit int64) (io.ReadCloser, error) {
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
	params := core.Params{
		"q": s,
	}

	var q = astral.NewQuery(c.consumerID, c.providerID, core.Query(methodSearch, params))

	conn, err := query.Route(ctx, c.mod.node, q)
	if err != nil {
		return nil, err
	}

	var results = make(chan *objects.SearchResult)

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	go func() {
		defer conn.Close()
		defer close(results)

		for {
			var sr = &objects.SearchResult{}
			_, err := sr.ReadFrom(conn)
			if err != nil {
				return
			}
			results <- sr
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

	params := core.Params{
		"size": strconv.FormatInt(int64(len(b)), 10),
	}

	var q = astral.NewQuery(c.consumerID, c.providerID, core.Query(methodPush, params))

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

func (c *Consumer) Describe(ctx *astral.Context, objectID *object.ID, _ *astral.Scope) (<-chan *objects.SourcedObject, error) {
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
			obj, _, err := c.mod.Blueprints().Read(conn, true)
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
