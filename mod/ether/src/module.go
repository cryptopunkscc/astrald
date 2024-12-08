package ether

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/ether"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/resources"
	"net"
	"strconv"
	"time"
)

const maxBroadcastSize = 1<<16 - 1

var _ ether.Module = &Module{}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets resources.Resources
	socket *net.UDPConn
}

func (mod *Module) Run(ctx context.Context) (err error) {
	if err = mod.setupSocket(ctx); err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		b, addr, err := mod.readBroadcast()
		if err != nil {
			mod.log.Errorv(2, "read broadcast: %v", err)
			continue
		}

		mod.Objects.Receive(&ether.EventBroadcastReceived{
			SourceID: b.Source,
			SourceIP: tcp.IP(addr.IP),
			Object:   b.Object,
		}, mod.node.Identity())

		err = mod.Objects.Receive(b.Object, b.Source)
		if err == nil {
			mod.log.Logv(2, "accepted object %v from %v", b.Object.ObjectType(), b.Source)
		} else {
			mod.log.Logv(2, "rejected object %v from %v", b.Object.ObjectType(), b.Source)
		}
	}
}

func (mod *Module) Push(object astral.Object, source *astral.Identity) (err error) {
	if source == nil {
		source = mod.node.Identity()
	}

	b := &ether.SignedBroadcast{
		Broadcast: ether.Broadcast{
			Object:    object,
			Timestamp: astral.Time(time.Now()),
			Source:    source,
		},
	}

	b.Signature, err = mod.Keys.Sign(source, b.Hash())
	if err != nil {
		return
	}

	if err = b.VerifySig(); err != nil {
		return
	}

	packet := &bytes.Buffer{}
	_, err = b.WriteTo(packet)
	if err != nil {
		return
	}

	err = mod.broadcast(packet.Bytes())

	return
}

// readBroadcast reads the next broadcast from the UDP socket
func (mod *Module) readBroadcast() (*ether.SignedBroadcast, *net.UDPAddr, error) {
	for {
		buf := make([]byte, maxBroadcastSize)

		n, srcAddr, err := mod.socket.ReadFromUDP(buf)
		if err != nil {
			return nil, nil, err
		}

		var r = objectReader{
			Reader:  bytes.NewReader(buf[:n]),
			objects: mod.Objects,
		}

		var b ether.SignedBroadcast
		_, err = b.ReadFrom(r)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid broadcast object from %s: %w", srcAddr, err)
		}

		// ignore our own broadcasts
		if b.Source.IsEqual(mod.node.Identity()) {
			continue
		}

		// verify signature
		if !ecdsa.VerifyASN1(b.Source.PublicKey().ToECDSA(), b.Hash(), b.Signature) {
			return nil, nil, fmt.Errorf("invalid object signature from %s", srcAddr)
		}

		return &b, srcAddr, nil
	}
}

// setupSocket sets up the UDP socket for broadcasts. If ctx is not nil, the socket will close when the context gets canceled.
func (mod *Module) setupSocket(ctx context.Context) (err error) {
	// resolve local address
	var localAddr *net.UDPAddr
	localAddr, err = net.ResolveUDPAddr("udp", ":"+strconv.Itoa(etherUDPPort))
	if err != nil {
		return
	}

	// bind the udp socket
	mod.socket, err = net.ListenUDP("udp", localAddr)

	// close the socket when the context is done
	if err == nil && ctx != nil {
		go func() {
			<-ctx.Done()
			mod.socket.Close()
		}()
	}
	return
}

func (mod *Module) broadcast(data []byte) error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	// go over all network interfaces
	for _, iface := range ifaces {
		if !isInterfaceEnabled(iface) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return err
		}

		// go over all addresses of the interface
		for _, addr := range addrs {
			broadcastIP, err := BroadcastAddr(addr)
			if err != nil {
				return err
			}

			if IsLinkLocal(broadcastIP) {
				continue
			}

			var broadcastAddr = net.UDPAddr{
				IP:   broadcastIP,
				Port: etherUDPPort,
			}

			_, err = mod.socket.WriteTo(data, &broadcastAddr)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isInterfaceEnabled(iface net.Interface) bool {
	return (iface.Flags&net.FlagUp != 0) &&
		(iface.Flags&net.FlagBroadcast != 0) &&
		(iface.Flags&net.FlagLoopback == 0)
}

func BroadcastAddr(addr net.Addr) (net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(addr.String())
	if err != nil {
		return nil, err
	}

	if len(ipnet.Mask) == net.IPv4len {
		ip = ip[12:]
	}

	broadIP := make(net.IP, len(ipnet.Mask))

	for i := 0; i < len(ipnet.Mask); i++ {
		broadIP[i] = ip[i] | ^ipnet.Mask[i]
	}

	return broadIP, nil
}

func IsLinkLocal(ip net.IP) bool {
	if ip := ip.To4(); ip != nil {
		return ip[0] == 169 && ip[1] == 254
	}
	return ip[0] == 0xfe && ip[1] == 0x80
}
