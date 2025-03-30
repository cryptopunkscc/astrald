package ether

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
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

func (mod *Module) Run(ctx *astral.Context) (err error) {
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
			objectID, err := astral.ResolveObjectID(b.Object)
			if err == nil {
				mod.log.Logv(2, "received from %v object %v (%v)", b.Source, b.Object.ObjectType(), objectID)
			}
		}
	}
}

// Push pushes an object to everyone (broadcast)
func (mod *Module) Push(object astral.Object, source *astral.Identity) (err error) {
	packet, err := mod.makePacket(object, source)
	if err != nil {
		return err
	}

	return mod.broadcast(packet)
}

// PushToIP pushes an object to a specific IP address (unicast)
func (mod *Module) PushToIP(ip tcp.IP, object astral.Object, source *astral.Identity) error {
	packet, err := mod.makePacket(object, source)
	if err != nil {
		return err
	}

	_, err = mod.writeToIP(ip, packet)
	return err
}

// readBroadcast reads the next broadcast from the UDP socket
func (mod *Module) readBroadcast() (*ether.SignedBroadcast, *net.UDPAddr, error) {
	for {
		buf := make([]byte, maxBroadcastSize)

		n, srcAddr, err := mod.socket.ReadFromUDP(buf)
		if err != nil {
			return nil, nil, err
		}

		var r = mod.Objects.Blueprints().Inject(bytes.NewReader(buf[:n]))

		var signed ether.SignedBroadcast
		_, err = signed.ReadFrom(r)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid broadcast object from %s: %w", srcAddr, err)
		}

		// ignore our own broadcasts
		if signed.Source.IsEqual(mod.node.Identity()) {
			continue
		}

		// verify signature
		if !ecdsa.VerifyASN1(signed.Source.PublicKey().ToECDSA(), signed.Hash(), signed.Signature) {
			return nil, nil, fmt.Errorf("invalid object signature from %s", srcAddr)
		}

		return &signed, srcAddr, nil
	}
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

			_, err = mod.writeToIP(tcp.IP(broadcastIP), data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (mod *Module) writeToIP(ip tcp.IP, data []byte) (n int, err error) {
	return mod.socket.WriteTo(data, &net.UDPAddr{
		IP:   net.IP(ip),
		Port: etherUDPPort,
	})
}

func (mod *Module) makePacket(object astral.Object, source *astral.Identity) (data []byte, err error) {
	if source == nil {
		source = mod.node.Identity()
	}

	signed := &ether.SignedBroadcast{
		Broadcast: ether.Broadcast{
			Object:    object,
			Timestamp: astral.Time(time.Now()),
			Source:    source,
		},
	}

	var hash = signed.Hash()

	signed.Signature, err = mod.Keys.SignASN1(source, hash)
	if err != nil {
		return
	}

	err = mod.Keys.VerifyASN1(source, hash, signed.Signature)
	if err != nil {
		return
	}

	packet := &bytes.Buffer{}
	_, err = signed.WriteTo(packet)
	if err != nil {
		return
	}

	return packet.Bytes(), nil
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
