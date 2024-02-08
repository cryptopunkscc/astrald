package presence

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/presence/proto"
	"github.com/cryptopunkscc/astrald/sig"
	"net"
	"slices"
	"strings"
	"time"
)

const announceInterval = 5 * time.Minute

type AnnounceService struct {
	*Module
	v chan bool
}

var _ presence.PendingAd = &PendingAd{}

type PendingAd struct {
	flags sig.Set[string]
}

func (ad *PendingAd) AddFlag(flag string) {
	ad.flags.Add(flag)
}

func (srv *AnnounceService) Run(ctx context.Context) error {
	srv.v = make(chan bool, 1)

	go srv.periodicAnnouncer(ctx)

	<-ctx.Done()
	return nil
}

func (srv *AnnounceService) periodicAnnouncer(ctx context.Context) {
	var enabled = srv.visible.Load()

	for {
		if enabled {
			if ad, err := srv.announceWithFlag(); err != nil {
				srv.log.Error("announce error: %s", err)
			} else {
				srv.log.Errorv(2, "announced presence with flags: %v", strings.Join(ad.Flags, ", "))
			}
		}

		select {
		case <-time.After(announceInterval - 5*time.Second): // broadcast 5s early to avoid presence timeout
		case e := <-srv.v:
			enabled = e

		case <-ctx.Done():
			return
		}
	}
}

func (srv *AnnounceService) announceWithFlag(flags ...string) (*proto.Ad, error) {
	ad, err := srv.signedAdWithFlags(flags...)
	if err != nil {
		return nil, err
	}

	buf, err := cslq.Marshal(ad)
	if err != nil {
		return nil, err
	}

	return ad, srv.bcastBytes(buf)
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
		if !slices.Contains(ad.Flags, f) {
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
	var pad = &PendingAd{}
	for _, f := range srv.outFilters.Clone() {
		f.OnPendingAd(pad)
	}
	return &proto.Ad{
		Identity:  srv.node.Identity(),
		Alias:     srv.myAlias(),
		ExpiresAt: time.Now().Add(announceInterval),
		Port:      srv.tcp.ListenPort(),
		Flags:     pad.flags.Clone(),
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
