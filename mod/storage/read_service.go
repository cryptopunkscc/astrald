package storage

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/storage/proto"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/tasks"
	"io"
	"sync"
	"sync/atomic"
)

var _ tasks.Runner = &ReadService{}

const ReadServiceName = "storage.read"

type ReadService struct {
	*Module
}

func (s *ReadService) Run(ctx context.Context) error {
	var queries = services.NewQueryChan(4)
	service, err := s.node.Services().Register(ctx, s.node.Identity(), ReadServiceName, queries.Push)
	if err != nil {
		return err
	}

	go func() {
		<-service.Done()
		close(queries)
	}()

	disco, err := modules.Find[*discovery.Module](s.node.Modules())
	if err == nil {
		disco.AddSource(s, s.node.Identity())
		go func() {
			<-ctx.Done()
			disco.RemoveSource(s)
		}()
	} else {
		s.log.Errorv(2, "can't regsiter service discovery source: %s", err)
	}

	for query := range queries {
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

func (s *ReadService) Discover(ctx context.Context, caller id.Identity, origin string) ([]discovery.ServiceEntry, error) {
	var list []discovery.ServiceEntry
	if s.DataAccessCountByIdentity(caller) > 0 {
		list = append(list, discovery.ServiceEntry{
			Name: ReadServiceName,
			Type: ReadServiceName,
		})
	}
	return list, nil
}

func (s *ReadService) handle(ctx context.Context, conn *services.Conn) error {
	defer conn.Close()
	return cslq.Invoke(conn, func(msg proto.MsgRead) error {
		var stream = proto.New(conn)

		if !s.CheckAccess(conn.RemoteIdentity(), msg.DataID) {
			s.log.Errorv(2, "access denied to %v to %v", conn.RemoteIdentity(), msg.DataID)
			return stream.WriteError(proto.ErrUnavailable)
		}

		source, err := s.findSource(ctx, msg, conn.RemoteIdentity())
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

func (s *ReadService) findSource(ctx context.Context, msg proto.MsgRead, identity id.Identity) (io.ReadCloser, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var ch = make(chan io.ReadCloser)

	var found atomic.Bool

	var wg sync.WaitGroup
	for source := range s.dataSources {
		source := source

		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := s.node.Services().Query(ctx, identity, source.Service, nil)

			if err != nil {
				switch {
				case errors.Is(err, context.Canceled):
				case errors.Is(err, context.DeadlineExceeded):
				default:
					s.RemoveDataSource(source)
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
