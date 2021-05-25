package lan

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/node/auth/id"
	"github.com/cryptopunkscc/astrald/node/net"
	"log"
	goNet "net"
	"strconv"
)

func (drv *driver) Scan(ctx context.Context) (<-chan *net.Ad, error) {
	udpAddr, err := goNet.ResolveUDPAddr("udp", ":"+strconv.Itoa(adPort))
	if err != nil {
		log.Println("cannot resolve udp address:", err)
		return nil, err
	}

	l, err := goNet.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("cannot scan:", err)
		return nil, err
	}
	log.Println("started scanning on lan")

	output := make(chan *net.Ad)

	go func() {
		buf := make([]byte, adLen)
		for {
			n, srcAddr, err := l.ReadFromUDP(buf)
			if err != nil {
				// Return on socket error
				log.Println("stopped scanning on lan")
				return
			}
			if n != adLen {
				// Ignore packets of invalid length
				continue
			}

			port := binary.BigEndian.Uint16(buf[0:2])
			pkData := buf[2:]
			srcIP, _, err := goNet.SplitHostPort(srcAddr.String())
			if err != nil {
				continue
			}

			finalAddr := fmt.Sprintf("%s:%d", srcIP, port)

			id, err := id.ParsePublicKey(pkData)

			// Ignore our own ads
			if id.String() == drv.localIdentity.String() {
				continue
			}

			ad := &net.Ad{
				Identity: id,
				Addr:     net.MakeAddr("lan", finalAddr),
			}

			output <- ad
		}
	}()

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	return output, nil
}

// ScanSeparately attempts to scan on each interface individually, so that we can have more control over which
// network interfaces to scan on.
// TODO: it fails to read broadcast data this way, not sure why
func (drv *driver) ScanSeparately(ctx context.Context) (<-chan *net.Ad, error) {
	ifaces, err := goNet.Interfaces()
	if err != nil {
		return nil, err
	}

	output := make(chan *net.Ad)

	for _, iface := range ifaces {
		if iface.Flags&goNet.FlagBroadcast != 0 {
			_ = scanInterface(ctx, iface, output)
		}
	}

	go func() {
		<-ctx.Done()
		close(output)
	}()

	return output, nil
}

func scanInterface(ctx context.Context, iface goNet.Interface, output chan<- *net.Ad) error {
	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}
	if len(addrs) < 1 {
		return errors.New("network interface has no addresses")
	}

	for _, addr := range addrs {
		_ = scanAddr(ctx, addr, output)
	}

	return nil
}

func scanAddr(ctx context.Context, addr net.Addr, output chan<- *net.Ad) error {
	ip, _ := net.SplitIPMask(addr.String())
	hostPort := goNet.JoinHostPort(ip, strconv.Itoa(adPort))

	udpAddr, err := goNet.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		log.Println("udp address error:", err)
		return err
	}

	l, err := goNet.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("cannot listen on", hostPort, ":", err)
		return err
	}
	log.Println("scanning on", hostPort)

	go func() {
		buf := make([]byte, adLen)
		for {
			n, srcAddr, err := l.ReadFromUDP(buf)

			fmt.Println("read", n, srcAddr, err)

			if err != nil {
				// Return on socket error
				return
			}
			if n != adLen {
				// Ignore packets of invalid length
				continue
			}

			port := binary.BigEndian.Uint16(buf[0:2])
			pkData := buf[2:]
			srcIP, _, err := goNet.SplitHostPort(srcAddr.String())
			if err != nil {
				continue
			}

			finalAddr := fmt.Sprintf("%s:%d", srcIP, port)

			id, err := id.ParsePublicKey(pkData)

			ad := &net.Ad{
				Identity: id,
				Addr:     net.MakeAddr("lan", finalAddr),
			}

			output <- ad
		}
	}()

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	return nil
}
