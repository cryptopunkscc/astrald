package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/ip"
	"log"
	"net"
	"strconv"
	"strings"
)

var _ infra.Network = &Inet{}

type Inet struct {
	config         Config
	listenPort     uint16
	publicAddrs    []infra.AddrSpec
	separateListen bool
}

func New(config Config) *Inet {
	inet := &Inet{
		config:      config,
		listenPort:  defaultPort,
		publicAddrs: make([]infra.AddrSpec, 0),
	}

	// Add public addresses
	for _, addrStr := range config.PublicAddr {
		addr, err := Parse(addrStr)
		if err != nil {
			log.Println("inet: parse error:", err)
			continue
		}

		inet.publicAddrs = append(inet.publicAddrs, infra.AddrSpec{
			Addr:   addr,
			Global: true,
		})
		log.Println("inet: added", addr)
	}

	return inet
}

func (inet Inet) Name() string {
	return NetworkName
}

func (inet Inet) Unpack(bytes []byte) (infra.Addr, error) {
	return Unpack(bytes)
}

func (inet Inet) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	a, ok := addr.(Addr)
	if !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	return Dial(ctx, a)
}

func (inet Inet) Listen(ctx context.Context) (<-chan infra.Conn, <-chan error) {
	if inet.separateListen {
		return inet.listenSeparately(ctx)
	}
	return inet.listenCombined(ctx)
}

func (inet Inet) Broadcast(payload []byte) error {
	return infra.ErrUnsupportedOperation
}

func (inet Inet) Scan(ctx context.Context) (<-chan infra.Broadcast, <-chan error) {
	return nil, singleErrChan(infra.ErrUnsupportedOperation)
}

func (inet Inet) Addresses() []infra.AddrSpec {
	list := make([]infra.AddrSpec, 0)

	ifaceAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	for _, a := range ifaceAddrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}

		ipv4 := ipnet.IP.To4()
		if ipv4 == nil {
			continue
		}

		if ipv4.IsLoopback() {
			continue
		}

		if ipv4.IsGlobalUnicast() || ipv4.IsPrivate() {
			list = append(list, infra.AddrSpec{
				Addr:   Addr{ip: ipv4, port: inet.listenPort},
				Global: ipv4.IsGlobalUnicast() && (!ipv4.IsPrivate()),
			})
		}
	}

	// Add custom addresses
	list = append(list, inet.publicAddrs...)

	return list
}

func (inet Inet) listenCombined(ctx context.Context) (<-chan infra.Conn, <-chan error) {
	output, errCh := make(chan infra.Conn), make(chan error, 1)

	go func() {
		defer close(output)
		defer close(errCh)

		hostPort := "0.0.0.0:" + strconv.Itoa(int(inet.listenPort))

		l, err := net.Listen("tcp", hostPort)
		if err != nil {
			errCh <- err
			return
		}

		log.Println("listen tcp", hostPort)

		go func() {
			<-ctx.Done()
			l.Close()
		}()

		for {
			conn, err := l.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					errCh <- err
				}
				return
			}

			output <- newConn(conn, false)
		}
	}()

	return output, errCh
}

func (inet Inet) listenSeparately(ctx context.Context) (<-chan infra.Conn, <-chan error) {
	output := make(chan infra.Conn)

	go func() {
		defer close(output)

		for ifaceName := range ip.Interfaces(ctx) {
			go func(ifaceName string) {
				for conn := range listenInterface(ctx, ifaceName) {
					output <- conn
				}
			}(ifaceName)
		}
	}()

	return output, nil
}

func singleErrChan(err error) <-chan error {
	ch := make(chan error, 1)
	defer close(ch)
	ch <- err
	return ch
}
