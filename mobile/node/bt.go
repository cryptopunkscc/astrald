package astralmobile

import (
	"context"
	"encoding/hex"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"io"
	"log"
)

// ==================================== Mobile bluetooth adapter ====================================

type BTClient interface {
	Address() string
	Listen() (BTPort, error)
	Dial(address []byte) (BTSocket, error)
}

type BTPort interface {
	Accept() (BTSocket, error)
	Close()
}

type BTSocket interface {
	io.WriteCloser      // Basic IO operations
	Read(Writer) error  // Gobind-hack, the replacement for standard io.Reader
	Outbound() bool     // Returns true if we are the active party, false otherwise
	LocalAddr() string  // Returns local network address if known, nil otherwise
	RemoteAddr() string // Returns the other party's network address if known, nil otherwise
}

// ==================================== Mobile bluetooth wrapper ====================================

func newBTAdapter(network BTClient) bt.Client {
	return &btClient{
		base:      bt.Bluetooth{},
		adapter:   network,
		addresses: []infra.AddrSpec{{infraAddr(network.Address()), false}},
	}
}

type btClient struct {
	base      bt.Bluetooth
	adapter   BTClient
	addresses []infra.AddrSpec
}

var _ infra.Network = &btClient{}
var _ infra.Unpacker = &btClient{}
var _ infra.Dialer = &btClient{}
var _ infra.Listener = &btClient{}
var _ infra.AddrLister = &btClient{}

func (b *btClient) Addresses() []infra.AddrSpec {
	return b.addresses
}

func (b *btClient) Name() string {
	return b.base.Name()
}

func (b *btClient) Unpack(network string, data []byte) (infra.Addr, error) {
	return b.base.Unpack(network, data)
}

func (b *btClient) Dial(_ context.Context, addr infra.Addr) (infra.Conn, error) {
	socket, err := b.adapter.Dial(addr.Pack())
	if err != nil {
		return nil, err
	}
	return newBTConn(socket), nil
}

func (b *btClient) Listen(ctx context.Context) (conn <-chan infra.Conn, err error) {
	// TODO BT Listen is not finished and not tested
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
	reader, writer := io.Pipe()
	go func() {
		err := socket.Read(writer)
		if err != nil {
			log.Println("BT socket read error", err)
		}
	}()
	return &btConn{
		BTSocket:   socket,
		localAddr:  infraAddr(socket.LocalAddr()),
		remoteAddr: infraAddr(socket.RemoteAddr()),
		reader:     reader,
		writer:     writer,
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
	reader     *io.PipeReader
	writer     *io.PipeWriter
}

var _ infra.Conn = &btConn{}

func (c *btConn) LocalAddr() infra.Addr {
	return c.localAddr
}

func (c *btConn) RemoteAddr() infra.Addr {
	return c.remoteAddr
}

func (c btConn) Read(bytes []byte) (l int, err error) {
	return c.reader.Read(bytes)
}

func (c btConn) Write(bytes []byte) (l int, err error) {
	return c.BTSocket.Write(bytes)
}

func formatHex(bytes []byte) string {
	str := hex.EncodeToString(bytes)
	f := ""
	for i, i2 := range str {
		f = f + string(i2)
		if i%2 != 0 {
			f = f + ":"
		}
	}
	return f
}
