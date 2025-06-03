package apphost

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
)

type Target struct {
	client   *Client
	targetID *astral.Identity
}

func (target *Target) Query(method string, args any) (conn astral.Conn, err error) {
	s, err := target.client.Session()
	if err != nil {
		return
	}

	var q = method
	if args != nil {
		params, err := query.Marshal(args)
		if err != nil {
			return nil, err
		}

		if len(params) > 0 {
			q = q + "?" + params
		}
	}

	return s.Query(target.client.GuestID, target.targetID, q)
}

func (target *Target) QueryChannel(method string, args any) (ch *astral.Channel, err error) {
	conn, err := target.Query(method, args)
	if err != nil {
		return
	}

	return astral.NewChannel(conn), nil
}

func (target *Target) ID() *astral.Identity {
	return target.targetID
}

func (target *Target) GetAlias(identity *astral.Identity) (string, error) {
	ch, err := target.QueryChannel("dir.get_alias", query.Args{
		"id": identity.String(),
	})
	if err != nil {
		return "", err
	}
	defer ch.Close()

	res, err := ch.Read()

	switch res := res.(type) {
	case *astral.String8:
		return string(*res), nil
	case *astral.ErrorMessage:
		return "", res
	default:
		return "", errors.New("unexpected response")
	}
}

func (target *Target) ResolveIdentity(name string) (*astral.Identity, error) {
	// try to parse the public key first
	if id, err := astral.IdentityFromString(name); err == nil {
		return id, nil
	}

	// then try using target's resolver
	ch, err := target.QueryChannel("dir.resolve", query.Args{
		"name": name,
	})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	obj, err := ch.Read()
	if err != nil {
		return nil, err
	}

	id, ok := obj.(*astral.Identity)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %s", obj.ObjectType())
	}

	return id, nil
}

func (target *Target) Describe(ctx context.Context, objectID *astral.ObjectID) (<-chan astral.Object, error) {
	ch, err := target.QueryChannel("objects.describe", query.Args{
		"id": objectID.String(),
	})
	if err != nil {
		return nil, err
	}

	res := make(chan astral.Object)
	done := make(chan struct{})

	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			ch.Close()
		}
	}()

	go func() {
		defer close(res)
		defer close(done)

		for {
			obj, err := ch.Read()
			if err != nil {
				return
			}

			res <- obj
		}
	}()

	return res, nil
}

func (target *Target) Push(ctx context.Context, object ...astral.Object) error {
	ch, err := target.QueryChannel("objects.push", nil)
	if err != nil {
		return err
	}

	var errs []error

	for idx, obj := range object {
		err := ch.Write(obj)
		if err != nil {
			errs = append(errs, fmt.Errorf("write at index %d: %w", idx, err))
			break
		}

		res, err := ch.Read()
		if err != nil {
			errs = append(errs, fmt.Errorf("read error: %w", err))
			break
		}

		switch res.(type) {
		case *astral.Ack:
			continue

		case *astral.ErrorMessage:
			errs = append(errs, fmt.Errorf("push at index %d: %w", idx, err))
		}
	}

	return errors.Join(errs...)
}

func (target *Target) Read(objectID *astral.ObjectID, offset, limit uint64) (io.ReadCloser, error) {
	conn, err := target.Query("objects.read", map[string]any{
		"id":     objectID,
		"offset": offset,
		"limit":  limit,
		"zone":   "dvn",
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}
