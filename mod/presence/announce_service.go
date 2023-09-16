package presence

import (
	"bytes"
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/presence/proto"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/inet"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/ip"
	"net"
	"time"
)

const announceInterval = 5 * time.Minute

type AnnounceService struct {
	*Module
	myAd *proto.Ad
}

func (srv *AnnounceService) Run(ctx context.Context) error {
	if !srv.config.Discoverable {
		return nil
	}

	srv.myAd = &proto.Ad{
		Identity: srv.node.Identity(),
		Port:     srv.getListenPort(),
		Flags:    proto.FlagDiscover,
	}
	srv.myAd.Alias, _ = srv.node.Tracker().GetAlias(srv.node.Identity())

	srv.log.Log("discoverable as %v", srv.node.Identity())

	if err := srv.broadcastPresence(); err != nil {
		return err
	}

	srv.myAd.Flags = 0

	for {
		select {
		case <-time.After(announceInterval):
			if err := srv.broadcastPresence(); err != nil {
				srv.log.Error("broadcast error: %s", err)
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (srv *AnnounceService) SendAdTo(dst *net.UDPAddr) error {
	if !srv.config.Discoverable {
		return ErrNotDiscoverable
	}

	return srv.sendAdTo(dst)
}

func (srv *AnnounceService) broadcastPresence() error {
	// prepare data
	data, err := srv.adData()
	if err != nil {
		return err
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, iface := range ifaces {
		if !isInterfaceEnabled(iface) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			broadcastIP, err := ip.BroadcastAddr(addr)
			if err != nil {
				return err
			}

			if ip.IsLinkLocal(broadcastIP) {
				continue
			}

			var broadcastAddr = net.UDPAddr{
				IP:   broadcastIP,
				Port: defaultPresencePort,
			}

			_, err = srv.socket.WriteTo(data, &broadcastAddr)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (srv *AnnounceService) adData() ([]byte, error) {
	var data = &bytes.Buffer{}
	if err := cslq.Encode(data, "v", srv.myAd); err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func (srv *AnnounceService) sendAdTo(dst *net.UDPAddr) error {
	data, err := srv.adData()
	if err != nil {
		return err
	}

	_, err = srv.socket.WriteTo(data, dst)
	return err
}

func (srv *AnnounceService) getListenPort() int {
	drv, ok := infra.GetDriver[*inet.Driver](srv.node.Infra(), inet.DriverName)
	if !ok {
		return -1
	}

	return drv.ListenPort()
}
