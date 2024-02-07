package presence

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/presence/proto"
	"net"
	"time"
)

const announceInterval = 5 * time.Minute

type AnnounceService struct {
	*Module
	myAd *proto.Ad
	v    chan bool
}

func (srv *AnnounceService) Run(ctx context.Context) error {
	srv.v = make(chan bool)

	srv.myAd = &proto.Ad{
		Identity: srv.node.Identity(),
		Port:     srv.tcp.ListenPort(),
		Flags:    []string{presence.DiscoverFlag},
	}

	srv.myAd.Alias, _ = srv.node.Tracker().GetAlias(srv.node.Identity())

	srv.myAd.Flags = nil

	go srv.periodicAnnouncer(ctx)

	<-ctx.Done()
	return nil
}

func (srv *AnnounceService) periodicAnnouncer(ctx context.Context) {
	var enabled = srv.visible.Load()

	for {
		if enabled {
			if err := srv.announceWithFlag(srv.flagsOnce.Clone()...); err != nil {
				srv.log.Error("broadcast error: %s", err)
			}
			srv.flagsOnce.Clear()
		}

		select {
		case <-time.After(announceInterval):
		case e := <-srv.v:
			enabled = e

		case <-ctx.Done():
			return
		}
	}
}

func (srv *AnnounceService) announceWithFlag(flags ...string) error {
	ad, err := srv.signedAdWithFlags(flags...)
	if err != nil {
		return err
	}

	buf, err := cslq.Marshal(ad)
	if err != nil {
		return err
	}

	return srv.bcastBytes(buf)
}

func (srv *AnnounceService) sendWithFlags(dst *net.UDPAddr, flags ...string) error {
	ad, err := srv.signedAdWithFlags(flags...)
	if err != nil {
		return err
	}

	buf, err := cslq.Marshal(ad)
	if err != nil {
		return err
	}

	_, err = srv.socket.WriteTo(buf, dst)
	return err
}

func (srv *AnnounceService) signedAdWithFlags(flags ...string) (*proto.Ad, error) {
	var err error
	var ad = srv.newAd()

	for _, f := range flags {
		if !srv.flags.Contains(f) {
			ad.Flags = append(ad.Flags, f)
		}
	}

	ad.Sig, err = srv.keys.Sign(srv.node.Identity(), ad.Hash())
	if err != nil {
		return nil, err
	}

	return ad, nil
}

func (srv *AnnounceService) newAd() *proto.Ad {
	return &proto.Ad{
		Identity: srv.node.Identity(),
		Alias:    srv.myAlias(),
		Port:     srv.tcp.ListenPort(),
		Flags:    srv.flags.Clone(),
	}
}

func (srv *AnnounceService) bcastBytes(data []byte) error {
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
