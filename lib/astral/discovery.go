package astral

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"io"
)

type DiscoveryHandler func(remoteID id.Identity) []ServiceInfo

type Discovery struct {
	*ApphostClient
}

type ServiceInfo struct {
	Name  string
	Type  string
	Extra []byte
}

func NewDiscovery(apphost *ApphostClient) *Discovery {
	return &Discovery{ApphostClient: apphost}
}

func (d *Discovery) RegisterSource(name string) (io.Closer, error) {
	rconn, err := d.Query(id.Identity{}, "services.register")
	if err != nil {
		return nil, err
	}

	conn := proto.New(rconn)
	conn.Encode(proto.MsgRegister{Service: name})

	err = conn.ReadError()
	if err != nil {
		return nil, err
	}

	return rconn, nil
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
				cslq.Encode(conn, "v", proto.ServiceEntry{
					Name:  s.Name,
					Type:  s.Type,
					Extra: s.Extra,
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
	var conn = proto.New(c)

	for err == nil {
		err = cslq.Invoke(conn, func(msg proto.ServiceEntry) error {
			list = append(list, ServiceInfo{
				Name:  msg.Name,
				Type:  msg.Type,
				Extra: msg.Extra,
			})
			return nil
		})
	}

	return list, nil
}
