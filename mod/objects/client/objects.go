package objects

import (
	"errors"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type Client struct {
	astral   *astrald.Client
	targetID *astral.Identity
}

var defaultClient *Client

func New(targetID *astral.Identity, astral *astrald.Client) *Client {
	if astral == nil {
		astral = astrald.DefaultClient()
	}
	return &Client{astral: astral, targetID: targetID}
}

func Default() *Client {
	if defaultClient == nil {
		defaultClient = New(nil, astrald.DefaultClient())
	}
	return defaultClient
}

func (client *Client) Create(ctx *astral.Context, repo string, alloc int) (objects.Writer, error) {
	// prepare arguments
	args := query.Args{}

	if alloc > 0 {
		args["alloc"] = alloc
	}
	if len(repo) > 0 {
		args["repo"] = repo
	}

	// send the query
	ch, err := client.queryCh(ctx, "objects.create", args)
	if err != nil {
		return nil, err
	}

	// handle response
	msg, err := ch.Receive()
	switch msg := msg.(type) {
	case *astral.Ack:
		return &writer{ch: ch}, nil

	case *astral.ErrorMessage:
		return nil, msg

	case nil:
		ch.Close()
		return nil, err
	}

	ch.Close()
	return nil, fmt.Errorf("unexpected message type: %s", msg.ObjectType())
}

func Create(ctx *astral.Context, repo string, alloc int) (objects.Writer, error) {
	return Default().Create(ctx, repo, alloc)
}

func (client *Client) Delete(ctx *astral.Context, objectID *astral.ObjectID, repo string) error {
	ch, err := client.queryCh(ctx, "objects.delete", query.Args{
		"id":   objectID,
		"repo": repo,
	})
	if err != nil {
		return err
	}

	msg, err := ch.Receive()
	switch msg.(type) {
	case nil:
		return err
	case *astral.Ack:
		return nil
	default:
		return errors.New("unexpected message type: " + msg.ObjectType())
	}
}

func Delete(ctx *astral.Context, objectID *astral.ObjectID, repo string) error {
	return Default().Delete(ctx, objectID, repo)
}

func (client *Client) Describe(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.DescribeResult, error) {
	ch, err := client.queryCh(ctx, "objects.describe", query.Args{
		"id": objectID.String(),
	}, channel.AllowUnparsed(true))
	if err != nil {
		return nil, err
	}

	res := make(chan *objects.DescribeResult)

	go func() {
		defer close(res)

		ch.Handle(ctx, func(o astral.Object) {
			switch o := o.(type) {
			case *objects.DescribeResult:
				res <- o

			case *astral.EOS:
				ch.Close()

			default:
				ch.Close()
			}
		})
	}()

	return res, nil
}

func Describe(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.DescribeResult, error) {
	return Default().Describe(ctx, objectID)
}

func (client *Client) NewMem(ctx *astral.Context, name string, size int64) error {
	// send the query
	ch, err := client.queryCh(ctx, "objects.new_mem", query.Args{
		"name": name,
		"size": size,
	})
	if err != nil {
		return err
	}

	// get the ack
	msg, err := ch.Receive()
	switch msg.(type) {
	case nil:
		return err
	case *astral.Ack:
		return nil
	default:
		return errors.New("unexpected message type: " + msg.ObjectType())
	}
}

func (client *Client) Probe(ctx *astral.Context, objectID *astral.ObjectID, repo string) (*objects.Probe, error) {
	// send the query
	ch, err := client.queryCh(ctx, "objects.probe", query.Args{
		"id":   objectID,
		"repo": repo,
	})
	if err != nil {
		return nil, err
	}

	// get the ack
	msg, err := ch.Receive()
	switch msg := msg.(type) {
	case nil:
		return nil, err
	case *objects.Probe:
		return msg, nil
	default:
		return nil, errors.New("unexpected message type: " + msg.ObjectType())
	}
}

func Probe(ctx *astral.Context, objectID *astral.ObjectID, repo string) (*objects.Probe, error) {
	return Default().Probe(ctx, objectID, repo)
}

func (client *Client) Repositories(ctx *astral.Context) (repos []*objects.RepositoryInfo, err error) {
	ch, err := client.queryCh(ctx, "objects.repositories", nil)
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	// collect repo names
	err = ch.Collect(func(object astral.Object) error {
		switch object := object.(type) {
		case *objects.RepositoryInfo:
			repos = append(repos, object)

		case *astral.EOS:
			return io.EOF

		default:
			return fmt.Errorf("unexpected object type: %s", object.ObjectType())
		}
		return nil
	})

	// EOF is expected
	if errors.Is(err, io.EOF) {
		err = nil
	}

	return
}

func (client *Client) Scan(ctx *astral.Context, repo string, follow bool) (<-chan *astral.ObjectID, error) {
	ch, err := client.queryCh(ctx, "objects.scan", query.Args{
		"repo":   repo,
		"follow": follow,
	})
	if err != nil {
		return nil, err
	}

	out := make(chan *astral.ObjectID)

	go func() {
		defer close(out)

		// collect the snapshot
		err := ch.Collect(func(o astral.Object) error {
			switch o := o.(type) {
			case *astral.ObjectID:
				select {
				case <-ctx.Done():
					return ctx.Err()
				case out <- o:
					return nil
				}

			case *astral.EOS:
				return io.EOF

			default:
				return fmt.Errorf("unexpected object type: %s", o.ObjectType())
			}
		})

		switch {
		case err == nil:
			return
		case errors.Is(err, io.EOF):
		default:
			return
		}

		if !follow {
			return
		}

		// send the separator
		select {
		case <-ctx.Done():
			return
		case out <- nil:
		}

		// handle updates
		ch.Handle(ctx, func(object astral.Object) {
			switch obj := object.(type) {
			case *astral.ObjectID:
				select {
				case <-ctx.Done():
					ch.Close()
				case out <- obj:
				}

			default:
				ch.Close()
			}
		})
	}()

	return out, nil
}

func Scan(ctx *astral.Context, repo string, follow bool) (<-chan *astral.ObjectID, error) {
	return Default().Scan(ctx, repo, follow)
}

func (client *Client) Search(ctx *astral.Context, q string) (<-chan *objects.SearchResult, error) {
	ch, err := client.queryCh(ctx, "objects.search", query.Args{
		"q": q,
	})
	if err != nil {
		return nil, err
	}

	res := make(chan *objects.SearchResult)
	go func() {
		defer close(res)

		ch.Handle(ctx, func(o astral.Object) {
			switch o := o.(type) {
			case *objects.SearchResult:
				res <- o

			case *astral.EOS:
				ch.Close()

			default:
				ch.Close()
			}
		})
	}()

	return res, nil
}

func Search(ctx *astral.Context, q string) (<-chan *objects.SearchResult, error) {
	return Default().Search(ctx, q)
}

func (client *Client) Push(ctx *astral.Context, object astral.Object) error {
	ch, err := client.queryCh(ctx, "objects.push", nil)
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.Send(object)
	if err != nil {
		return err
	}

	res, err := ch.Receive()
	switch res := res.(type) {
	case *astral.Bool:
		if *res {
			return nil
		} else {
			return errors.New("object rejected")
		}

	case nil:
		return err

	default:
		return fmt.Errorf("unexpected response type: %s", res.ObjectType())
	}
}

func Push(ctx *astral.Context, object astral.Object) error {
	return Default().Push(ctx, object)
}

func (client *Client) Read(ctx *astral.Context, objectID *astral.ObjectID, offset, limit int64) (io.ReadCloser, error) {
	return client.query(ctx, "objects.read", query.Args{
		"id":     objectID,
		"offset": offset,
		"limit":  limit,
		"zone":   "dvn",
	})
}

func Read(ctx *astral.Context, objectID *astral.ObjectID, offset, limit int64) (io.ReadCloser, error) {
	return Default().Read(ctx, objectID, offset, limit)
}

func (client *Client) query(ctx *astral.Context, method string, args any) (astral.Conn, error) {
	return client.astral.WithTarget(client.targetID).Query(ctx, method, args)
}

func (client *Client) queryCh(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return client.astral.WithTarget(client.targetID).QueryChannel(ctx, method, args, cfg...)
}

// Deprecated: use Probe instead.
func (client *Client) GetType(ctx *astral.Context, objectID *astral.ObjectID) (string, error) {
	ch, err := client.queryCh(ctx, "objects.get_type", query.Args{
		"id": objectID,
	})
	if err != nil {
		return "", err
	}
	defer ch.Close()

	res, err := ch.Receive()
	if err != nil {
		return "", err
	}

	switch res := res.(type) {
	case *astral.String8:
		return string(*res), nil

	case *astral.ErrorMessage:
		return "", res

	default:
		return "", errors.New("protocol error: unexpected object type")
	}
}
