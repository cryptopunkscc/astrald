package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"io"
	"log"
)

// ==================================== Mobile bluetooth adapter ====================================

// Bluetooth gomobile binding interface for bluetooth have to be implemented and injected from android sources.
type Bluetooth interface {
	Dial(address []byte) (BluetoothSocket, error)
}

type BluetoothSocket interface {
	io.WriteCloser      // Basic IO operations
	Read(Writer) error  // Gobind-hack, the replacement for standard io.Reader
	Outbound() bool     // Returns true if we are the active party, false otherwise
	LocalAddr() string  // Returns local network address if known, nil otherwise
	RemoteAddr() string // Returns the other party's network address if known, nil otherwise
}

type Writer io.Writer

// ==================================== Mobile bluetooth wrapper ====================================

func newBluetoothAdapter(client Bluetooth) bt.Client {
	return &bluetooth{
		base:   bt.Bluetooth{},
		native: client,
	}
}

// bluetooth adapts native android bluetooth to astral infra interfaces.
type bluetooth struct {
	base   bt.Bluetooth
	native Bluetooth
}

var _ bt.Client = &bluetooth{}

func (b *bluetooth) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (b *bluetooth) Name() string {
	return b.base.Name()
}

func (b *bluetooth) Dial(_ context.Context, addr bt.Addr) (infra.Conn, error) {
	socket, err := b.native.Dial(addr.Pack())
	if err != nil {
		return nil, err
	}
	return newBluetoothSocket(socket), nil
}

func newBluetoothSocket(socket BluetoothSocket) *bluetoothSocket {
	reader, writer := io.Pipe()
	go func() {
		err := socket.Read(writer)
		writer.Close()
		reader.Close()
		socket.Close()
		if err != nil {
			log.Println("BT socket read error", err)
		}
	}()
	return &bluetoothSocket{
		BluetoothSocket: socket,
		localAddr:       nil,
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
