package lan

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/cryptopunkscc/astrald/node/net"
	"log"
	goNet "net"
	"strconv"
	"time"
)

// A very crude local network discovery until I rework the net package

const adPort = 8829
const broadcastIP = "192.168.1.255"
const adLen = 35
const adInterval = time.Second

func (drv *driver) Advertise(ctx context.Context) error {
	ifaces, err := goNet.Interfaces()
	if err != nil {
		return err
	}

	for _, iface := range ifaces {
		if iface.Flags&goNet.FlagBroadcast != 0 {
			drv.advertiseOnIface(ctx, iface)
		}
	}

	return nil
}

func (drv *driver) advertiseOnIface(ctx context.Context, iface goNet.Interface) error {
	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}
	if len(addrs) < 1 {
		return errors.New("network interface has no addresses")
	}

	for _, addr := range addrs {
		_ = drv.advertiseOnAddr(ctx, addr)
	}

	return nil
}

func (drv *driver) advertiseOnAddr(ctx context.Context, addr net.Addr) error {
	broadAddr, err := broadcastAddr(addr)
	if err != nil {
		log.Println("cannot advertise to", broadAddr.String(), "because of parse error:", err)
		return err
	}

	broadStr := goNet.JoinHostPort(broadAddr.String(), strconv.Itoa(adPort))

	conn, err := goNet.Dial("udp", broadStr)
	if err != nil {
		return err
	}

	log.Println("advertising to", broadStr)

	go func() {
		ad := drv.makeAd()
		defer conn.Close()
		for {
			_, err := conn.Write(ad)
			if err != nil {
				log.Println("error writing ad:", err)
				return
			}

			// log.Println("ad sent!")

			select {
			case <-time.After(adInterval):
				continue
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (drv *driver) makeAd() []byte {
	var ad = make([]byte, adLen)
	binary.BigEndian.PutUint16(ad[0:2], drv.port)
	idBytes, _ := hex.DecodeString(drv.localIdentity.String())
	copy(ad[2:], idBytes)
	return ad
}

func broadcastAddr(addr net.Addr) (net.Addr, error) {
	ip, ipnet, err := goNet.ParseCIDR(addr.String())
	if err != nil {
		return nil, err
	}

	if len(ipnet.Mask) == goNet.IPv4len {
		ip = ip[12:]
	}

	broadIP := make(goNet.IP, len(ipnet.Mask))

	for i := 0; i < len(ipnet.Mask); i++ {
		broadIP[i] = ip[i] | ^ipnet.Mask[i]
	}

	return net.MakeAddr(addr.Network(), broadIP.String()), nil
}
