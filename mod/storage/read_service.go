package storage

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/storage/proto"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/tasks"
	"io"
	"sync"
	"sync/atomic"
)

var _ tasks.Runner = &ReadService{}

type ReadService struct {
	*Module
}

func (s *ReadService) Run(ctx context.Context) error {
	srv, err := s.node.Services().RegisterContext(ctx, "storage.read")
	if err != nil {
		return err
	}

	for query := range srv.Queries() {
		conn, err := query.Accept()
		if err != nil {
			continue
		}

		go func() {
			if err := s.handle(ctx, conn); err != nil {
				s.log.Errorv(0, "read(): %s", err)
			}
		}()
	}

	return nil
}

func (s *ReadService) handle(ctx context.Context, conn *services.Conn) error {
	defer conn.Close()
	return cslq.Invoke(conn, func(msg proto.MsgRead) error {
		var stream = proto.New(conn)

		source, err := s.findSource(ctx, msg)
		if err != nil {
			return stream.WriteError(proto.ErrUnavailable)
		}
		defer source.Close()

		if err := stream.WriteError(nil); err != nil {
			return err
		}

		_, err = io.Copy(conn, source)
		return err
	})
}

func (s *ReadService) findSource(ctx context.Context, msg proto.MsgRead) (io.ReadCloser, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var ch = make(chan io.ReadCloser)

	var found atomic.Bool

	var wg sync.WaitGroup
	for source := range s.sources {
		source := source

		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := s.node.Query(ctx, s.node.Identity(), source.Service)
			if err != nil {
				switch {
				case errors.Is(err, context.Canceled):
				case errors.Is(err, context.DeadlineExceeded):
				default:
					s.RemoveSource(source)
				}
				return
			}

			var stream = proto.New(conn)

			if err := stream.Encode(msg); err != nil {
				return
			}

			err = stream.ReadError()
			if err != nil {
				return
			}

			if found.CompareAndSwap(false, true) {
				ch <- conn
				return
			}

			conn.Close()
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	if source, ok := <-ch; ok {
		return source, nil
	}

	return nil, errors.New("no source found")
}
