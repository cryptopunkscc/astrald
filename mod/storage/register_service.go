package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/storage/proto"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ tasks.Runner = &RegisterService{}

type RegisterService struct {
	*Module
}

func (m *RegisterService) Run(ctx context.Context) error {
	srv, err := m.node.Services().RegisterContext(ctx, "storage.register")
	if err != nil {
		return err
	}

	go func() {
		host, err := modules.Find[*apphost.Module](m.node.Modules())
		if err != nil {
			m.log.Logv(2, "apphost module not found, data sources will not be launched")
			return
		}
		for _, source := range m.config.Sources {
			source := source
			go func() {
				m.log.Infov(1, "launching: %s", source)
				err := host.LaunchRaw(source.Exec, source.Args...)
				m.log.Error("source %s ended with error: %s", source, err)
			}()
		}
	}()

	for query := range srv.Queries() {
		conn, err := query.Accept()
		if err != nil {
			continue
		}

		go func() {
			if err := m.handle(ctx, conn); err != nil {
				m.log.Errorv(0, "register(): %s", err)
			}
		}()
	}

	return nil
}

func (m *RegisterService) handle(ctx context.Context, conn *services.Conn) error {
	defer conn.Close()
	return cslq.Invoke(conn, func(msg proto.MsgRegisterSource) error {
		source := &Source{
			Service: msg.Service,
		}

		m.AddSource(source)

		return cslq.Encode(conn, "c", proto.Success)
	})
}
