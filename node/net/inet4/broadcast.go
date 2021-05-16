package inet4

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/node/auth"
	"github.com/cryptopunkscc/astrald/node/net"
	goNet "net"
	"strings"
)

// A very crude local network discovery until I rework the net package

const adPort = 8829
const broadcastIP = "192.168.1.255"
const adLen = 35

type Broadcast struct {
	pc    goNet.PacketConn
	baddr *goNet.UDPAddr
	msg   []byte
}

type Discovery struct {
	auth.Identity
	net.Endpoint
}

func NewBroadcast(localIdentity *auth.ECIdentity, localPort uint16) (*Broadcast, error) {
	var err error
	ad := &Broadcast{}

	ad.pc, err = goNet.ListenPacket("udp4", fmt.Sprintf(":%d", adPort))
	if err != nil {
		return nil, err
	}

	ad.baddr, err = goNet.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", broadcastIP, adPort))
	if err != nil {
		return nil, err
	}

	var buf = make([]byte, adLen)

	binary.BigEndian.PutUint16(buf[0:2], localPort)

	copy(buf[2:], localIdentity.PublicKey().SerializeCompressed())

	ad.msg = buf

	return ad, nil
}

func (ad *Broadcast) Advertise() error {
	_, err := ad.pc.WriteTo(ad.msg, ad.baddr)

	return err
}

func (ad *Broadcast) Scan() (*Discovery, error) {
	buf := make([]byte, adLen)
	n, addr, err := ad.pc.ReadFrom(buf)
	if err != nil {
		return nil, err
	}
	if n != 35 {
		return nil, errors.New("invalid packet")
	}

	port := binary.BigEndian.Uint16(buf[0:2])

	pkData := buf[2:]

	ip := strings.Split(addr.String(), ":")[0]

	finalAddr := fmt.Sprintf("%s:%d", ip, port)

	id, err := auth.ParsePublicKey(pkData)

	d := &Discovery{
		Identity: id,
		Endpoint: net.Endpoint{
			Net:     "inet4",
			Address: finalAddr,
		},
	}
	return d, nil
}
