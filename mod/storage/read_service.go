package storage

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/sdp"
	"github.com/cryptopunkscc/astrald/mod/storage/rpc"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/modules"
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

func (service *ReadService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		service.handle(ctx, conn, hints)
	})
}

func (service *ReadService) Run(ctx context.Context) error {
	err := service.node.AddRoute(ReadServiceName, service)
	if err != nil {
		return err
	}
	defer service.node.RemoveRoute(ReadServiceName)

	disco, err := modules.Find[*sdp.Module](service.node.Modules())
	if err == nil {
		disco.AddSource(service)
		defer disco.RemoveSource(service)
	} else {
		service.log.Errorv(2, "can't regsiter service discovery source: %s", err)
	}

	<-ctx.Done()

	return nil
}

func (service *ReadService) Discover(ctx context.Context, caller id.Identity, origin string) ([]sdp.ServiceEntry, error) {
	var list []sdp.ServiceEntry
	if service.DataAccessCountByIdentity(caller) > 0 {
		list = append(list, sdp.ServiceEntry{
			Name: ReadServiceName,
			Type: ReadServiceName,
		})
	}
	return list, nil
}

func (service *ReadService) handle(ctx context.Context, conn net.SecureConn, hints net.Hints) error {
	return cslq.Invoke(conn, func(msg rpc.MsgRead) error {
		var session = rpc.New(conn)

		if !service.CheckAccess(conn.RemoteIdentity(), msg.DataID) {
			service.log.Errorv(2, "access denied to %v to %v", conn.RemoteIdentity(), msg.DataID)
			return session.EncodeErr(rpc.ErrUnavailable)
		}

		source, err := service.findSource(ctx, msg, conn.RemoteIdentity(), hints)
		if err != nil {
			return session.EncodeErr(rpc.ErrUnavailable)
		}
		defer source.Close()

		if err := session.EncodeErr(nil); err != nil {
			return err
		}

		_, err = io.Copy(conn, source)
		conn.Close()
		return err
	})
}

func (service *ReadService) findSource(ctx context.Context, msg rpc.MsgRead, caller id.Identity, hints net.Hints) (io.ReadCloser, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var ch = make(chan io.ReadCloser)

	var found atomic.Bool

	var wg sync.WaitGroup
	for source := range service.dataSources {
		source := source

		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := net.RouteWithHints(
				ctx,
				service.node.Router(),
				net.NewQuery(caller, source.Identity, source.Service),
				hints,
			)

			if err != nil {
				switch {
				case errors.Is(err, context.Canceled):
				case errors.Is(err, context.DeadlineExceeded):
				default:
					service.RemoveDataSource(source)
				}
				return
			}

			_, err = rpc.New(conn).Read(msg.DataID, int(msg.Start), int(msg.Len))

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
