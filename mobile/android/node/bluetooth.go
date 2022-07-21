package astral

import (
	"context"
	"encoding/hex"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"io"
	"log"
)

// ==================================== Mobile bluetooth adapter ====================================

// Bluetooth gomobile binding interface for bluetooth have to be implemented and injected from android sources.
type Bluetooth interface {
	Address() string
	Listen() (BluetoothPort, error)
	Dial(address []byte) (BluetoothSocket, error)
}

type BluetoothPort interface {
	Accept() (BluetoothSocket, error)
	Close()
}

type BluetoothSocket interface {
	io.WriteCloser      // Basic IO operations
	Read(Writer) error  // Gobind-hack, the replacement for standard io.Reader
	Outbound() bool     // Returns true if we are the active party, false otherwise
	LocalAddr() string  // Returns local network address if known, nil otherwise
	RemoteAddr() string // Returns the other party's network address if known, nil otherwise
}

// ==================================== Mobile bluetooth wrapper ====================================

func newBluetoothAdapter(client Bluetooth) bt.Client {
	return &bluetooth{
		base:      bt.Bluetooth{},
		native:    client,
		addresses: []infra.AddrSpec{{infraAddr(client.Address()), false}},
	}
}

// bluetooth adapts native android bluetooth to astral infra interfaces.
type bluetooth struct {
	base      bt.Bluetooth
	native    Bluetooth
	addresses []infra.AddrSpec
}

var _ infra.Network = &bluetooth{}
var _ infra.Unpacker = &bluetooth{}
var _ infra.Dialer = &bluetooth{}
var _ infra.Listener = &bluetooth{}
var _ infra.AddrLister = &bluetooth{}

func (b *bluetooth) Addresses() []infra.AddrSpec {
	return b.addresses
}

func (b *bluetooth) Name() string {
	return b.base.Name()
}

func (b *bluetooth) Unpack(network string, data []byte) (infra.Addr, error) {
	return b.base.Unpack(network, data)
}

func (b *bluetooth) Dial(_ context.Context, addr infra.Addr) (infra.Conn, error) {
	socket, err := b.native.Dial(addr.Pack())
	if err != nil {
		return nil, err
	}
	return newBluetoothSocket(socket), nil
}

func (b *bluetooth) Listen(ctx context.Context) (conn <-chan infra.Conn, err error) {
	// TODO BT Listen is not finished and not tested
	c := make(chan infra.Conn)
	port, err := b.native.Listen()
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
			c <- newBluetoothSocket(socket)
		}
	}()
	conn = c
	return
}

func newBluetoothSocket(socket BluetoothSocket) *bluetoothSocket {
	reader, writer := io.Pipe()
	go func() {
		err := socket.Read(writer)
		if err != nil {
			log.Println("BT socket read error", err)
		}
	}()
	return &bluetoothSocket{
		BluetoothSocket: socket,
		localAddr:       infraAddr(socket.LocalAddr()),
		remoteAddr:      infraAddr(socket.RemoteAddr()),
		reader:          reader,
		writer:          writer,
	}
}

func infraAddr(addr string) infra.Addr {
	parsed, err := bt.Parse(addr)
	if err != nil {
		log.Println("Cannot parse bt address", addr, err)
	}
	return parsed
}

type bluetoothSocket struct {
	BluetoothSocket
	localAddr  infra.Addr
	remoteAddr infra.Addr
	reader     *io.PipeReader
	writer     *io.PipeWriter
}

var _ infra.Conn = &bluetoothSocket{}

func (c *bluetoothSocket) LocalAddr() infra.Addr {
	return c.localAddr
}

func (c *bluetoothSocket) RemoteAddr() infra.Addr {
	return c.remoteAddr
}

func (c bluetoothSocket) Read(bytes []byte) (l int, err error) {
	return c.reader.Read(bytes)
}

func (c bluetoothSocket) Write(bytes []byte) (l int, err error) {
	return c.BluetoothSocket.Write(bytes)
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
