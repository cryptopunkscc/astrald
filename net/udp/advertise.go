package udp

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/net/ip"
	"log"
	go_net "net"
	"strconv"
	"time"
)

// A very crude local network discovery until I rework the net package

const adPort = 8829
const adLen = 35
const adInterval = time.Second

//TODO: rename to Announce?
func (drv *driver) Advertise(ctx context.Context, id string) error {
	ifaces, err := go_net.Interfaces()
	if err != nil {
		return err
	}

	for _, iface := range ifaces {
		if iface.Flags&go_net.FlagBroadcast != 0 {
			drv.advertiseOnIface(ctx, iface, id)
		}
	}

	return nil
}

func (drv *driver) advertiseOnIface(ctx context.Context, iface go_net.Interface, id string) error {
	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}
	if len(addrs) < 1 {
		return errors.New("network interface has no addresses")
	}

	for _, addr := range addrs {
		_ = drv.advertiseOnAddr(ctx, addr, id)
	}

	return nil
}

func (drv *driver) advertiseOnAddr(ctx context.Context, addr net.Addr, id string) error {
	broadAddr, err := ip.BroadcastAddr(addr)
	if err != nil {
		log.Println("cannot advertise to", broadAddr.String(), "because of parse error:", err)
		return err
	}

	broadStr := go_net.JoinHostPort(broadAddr.String(), strconv.Itoa(adPort))

	conn, err := go_net.Dial("udp", broadStr)
	if err != nil {
		return err
	}

	log.Println("advertising to", broadStr)

	go func() {
		ad := drv.makeAd(id)
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

func (drv *driver) makeAd(id string) []byte {
	var ad = make([]byte, adLen)
	binary.BigEndian.PutUint16(ad[0:2], 1791)
	idBytes, _ := hex.DecodeString(id)
	copy(ad[2:], idBytes)
	return ad
}
