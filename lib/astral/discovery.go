package astral

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/discovery/rpc"
	"io"
)

type DiscoveryHandler func(remoteID id.Identity) []ServiceInfo

type Discovery struct {
	*ApphostClient
}

type ServiceInfo struct {
	Identity id.Identity
	Name     string
	Type     string
	Extra    []byte
}

func NewDiscovery(apphost *ApphostClient) *Discovery {
	return &Discovery{ApphostClient: apphost}
}

func (d *Discovery) RegisterSource(name string) (io.Closer, error) {
	conn, err := d.Query(id.Identity{}, "services.register")
	if err != nil {
		return nil, err
	}

	session := rpc.New(conn)

	err = session.Register(name)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (d *Discovery) RegisterHandler(handler DiscoveryHandler) (io.Closer, error) {
	var name = "services.discovery.source." + randomName(16)
	l, err := d.Register(name)
	if err != nil {
		return nil, err
	}

	if _, err := d.RegisterSource(name); err != nil {
		l.Close()
		return nil, err
	}

	go func() {
		for conn := range l.AcceptAll() {
			conn := conn.(*Conn)

			services := handler(conn.RemoteIdentity())

			for _, s := range services {
				cslq.Encode(conn, "v", rpc.ServiceEntry{
					Identity: s.Identity,
					Name:     s.Name,
					Type:     s.Type,
					Extra:    s.Extra,
				})
			}

			conn.Close()
		}
	}()

	return l, nil
}

func (d *Discovery) Discover(identity id.Identity) ([]ServiceInfo, error) {
	c, err := d.Query(identity, "services.discover")
	if err != nil {
		return nil, err
	}

	var list []ServiceInfo

	for err == nil {
		err = cslq.Invoke(c, func(msg rpc.ServiceEntry) error {
			list = append(list, ServiceInfo{
				Identity: msg.Identity,
				Name:     msg.Name,
				Type:     msg.Type,
				Extra:    msg.Extra,
			})
			return nil
		})
	}

	return list, nil
}
