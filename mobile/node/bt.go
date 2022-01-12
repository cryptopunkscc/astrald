package astralmobile

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"io"
	"log"
)

// ==================================== Mobile bluetooth adapter ====================================

type BTAdapter interface {
	Address() string
	Listen() (BTPort, error)
	Dial(address []byte) (BTSocket, error)
}

type BTPort interface {
	Accept() (BTSocket, error)
	Close()
}

type BTSocket interface {
	io.ReadWriteCloser  // Basic IO operations
	Outbound() bool     // Returns true if we are the active party, false otherwise
	LocalAddr() string  // Returns local network address if known, nil otherwise
	RemoteAddr() string // Returns the other party's network address if known, nil otherwise
}

// ==================================== Mobile bluetooth wrapper ====================================

func newBTWrapper(adapter BTAdapter) bt.BluetoothAdapter {
	return &btWrapper{
		base:      bt.Bluetooth{},
		adapter:   adapter,
		addresses: []infra.AddrSpec{{infraAddr(adapter.Address()), false}},
	}
}

type btWrapper struct {
	base      bt.Bluetooth
	adapter   BTAdapter
	addresses []infra.AddrSpec
}

func (b *btWrapper) Addresses() []infra.AddrSpec {
	return b.addresses
}

var _ infra.Network = &btWrapper{}
var _ infra.Unpacker = &btWrapper{}
var _ infra.Dialer = &btWrapper{}
var _ infra.Listener = &btWrapper{}
var _ infra.AddrLister = &btWrapper{}

func (b *btWrapper) Name() string {
	return b.base.Name()
}

func (b *btWrapper) Unpack(network string, data []byte) (infra.Addr, error) {
	return b.base.Unpack(network, data)
}

func (b *btWrapper) Dial(_ context.Context, addr infra.Addr) (infra.Conn, error) {
	socket, err := b.adapter.Dial(addr.Pack())
	if err != nil {
		return nil, err
	}
	return newBTConn(socket), nil
}

func (b *btWrapper) Listen(ctx context.Context) (conn <-chan infra.Conn, err error) {
	c := make(chan infra.Conn)
	port, err := b.adapter.Listen()
	if err != nil {
		return
	}
	running := true
	go func() {
		<-ctx.Done()
		running = false
		port.Close()
		close(c)
	}()
	go func() {
		log.Printf("(%s) listen %s\n", bt.NetworkName, b.addresses[0])
		for running {
			socket, err := port.Accept()
			if err != nil {
				log.Println("Cannot accept bt connection", err)
				continue
			}
			c <- newBTConn(socket)
		}
	}()
	conn = c
	return
}

func newBTConn(socket BTSocket) *btConn {
	return &btConn{
		BTSocket:   socket,
		localAddr:  infraAddr(socket.LocalAddr()),
		remoteAddr: infraAddr(socket.RemoteAddr()),
	}
}

func infraAddr(addr string) infra.Addr {
	parsed, err := bt.Parse(addr)
	if err != nil {
		log.Println("Cannot parse bt address", addr, err)
	}
	return parsed
}

type btConn struct {
	BTSocket
	localAddr  infra.Addr
	remoteAddr infra.Addr
}

var _ infra.Conn = &btConn{}

func (c *btConn) LocalAddr() infra.Addr {
	return c.localAddr
}

func (c *btConn) RemoteAddr() infra.Addr {
	return c.remoteAddr
}
