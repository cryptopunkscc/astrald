package route

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/jxskiss/base62"
	"io"
	"strings"
)

type Route struct {
	Identity  id.Identity
	Addresses []infra.Addr
}

func New(identity id.Identity) *Route {
	return &Route{
		Identity:  identity,
		Addresses: make([]infra.Addr, 0),
	}
}

const routePrefix = "node1"

func (route *Route) Add(addr infra.Addr) {
	for _, a := range route.Addresses {
		if infra.AddrEqual(a, addr) {
			return
		}
	}
	route.Addresses = append(route.Addresses, addr)
}

func (route Route) Pack() []byte {
	buf := &bytes.Buffer{}
	_ = Write(buf, &route)
	return buf.Bytes()
}

func (route Route) String() string {
	return routePrefix + base62.EncodeToString(route.Pack())
}

func (route Route) Each() <-chan infra.Addr {
	ch := make(chan infra.Addr, len(route.Addresses))
	defer close(ch)

	for _, addr := range route.Addresses {
		ch <- addr
	}

	return ch
}

func Parse(s string) (*Route, error) {
	str := strings.TrimPrefix(s, routePrefix)

	data, err := base62.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return Unpack(data)
}

func Unpack(data []byte) (*Route, error) {
	return Read(bytes.NewReader(data))
}

func Write(w io.Writer, route *Route) error {
	err := enc.WriteIdentity(w, route.Identity)
	if err != nil {
		return err
	}

	addrs := route.Addresses[:]
	if len(addrs) > 255 {
		addrs = addrs[:255]
	}

	err = enc.Write(w, uint8(len(addrs)))
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if err := enc.WriteAddr(w, addr); err != nil {
			return nil
		}
	}

	return nil
}

func Read(r io.Reader) (*Route, error) {
	_id, err := enc.ReadIdentity(r)
	if err != nil {
		return nil, err
	}

	count, err := enc.ReadUint8(r)
	if err != nil {
		return nil, err
	}

	addrs := make([]infra.Addr, 0, count)
	for i := 0; i < int(count); i++ {
		addr, err := enc.ReadAddr(r)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}

	return &Route{
		Identity:  _id,
		Addresses: addrs,
	}, nil
}
