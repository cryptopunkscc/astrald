package presence

import (
	"bytes"
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/presence/proto"
	"net"
	"strconv"
	"time"
)

const PresenceTimeout = 15 * time.Minute

type DiscoverService struct {
	*Module
	cache map[string]*Ad
}

func NewDiscoverService(module *Module) *DiscoverService {
	return &DiscoverService{
		Module: module,
		cache:  make(map[string]*Ad),
	}
}

func (srv *DiscoverService) Run(ctx context.Context) error {
	for {
		ad, err := srv.readAd()
		if err != nil {
			return err
		}

		srv.save(ad)

		srv.log.Logv(
			2,
			"received an ad from %v endpoint %v",
			ad.Identity,
			ad.Endpoint,
		)

		srv.events.Emit(EventAdReceived{ad})

		if ad.DiscoverFlag() {
			srv.Announce.SendAdTo(ad.UDPAddr)
		}

		if srv.config.AutoAdd {
			_ = srv.node.Tracker().AddEndpoint(ad.Identity, ad.Endpoint)
		}

		if srv.config.TrustAliases && ad.Alias != "" {
			if _, err := srv.node.Tracker().GetAlias(ad.Identity); err != nil {
				srv.node.Tracker().SetAlias(ad.Identity, ad.Alias)
				srv.log.Info("alias set for %v (%v)", ad.Identity, ad.Identity.Fingerprint())
			}
		}
	}
}

func (srv *DiscoverService) RecentAds() []*Ad {
	var res = make([]*Ad, 0, len(srv.cache))
	for _, p := range srv.cache {
		res = append(res, p)
	}
	return res
}

func (srv *DiscoverService) readAd() (*Ad, error) {
	for {
		buf := make([]byte, 1024)

		n, srcAddr, err := srv.socket.ReadFromUDP(buf)
		if err != nil {
			return nil, err
		}

		var msg proto.Ad
		err = cslq.Decode(bytes.NewReader(buf[:n]), "v", &msg)
		if err != nil {
			srv.log.Errorv(2, "received an invalid ad from %v: %v", srcAddr, err)
			continue
		}

		// ignore our own ad
		if msg.Identity.IsEqual(srv.node.Identity()) {
			continue
		}

		hostPort := net.JoinHostPort(srcAddr.IP.String(), strconv.Itoa(msg.Port))

		endpoint, err := srv.tcp.Parse("tcp", hostPort)
		if err != nil {
			panic(err)
		}

		return &Ad{
			UDPAddr:   srcAddr,
			Identity:  msg.Identity,
			Alias:     msg.Alias,
			Endpoint:  endpoint,
			Timestamp: time.Now(),
			Flags:     int(msg.Flags),
		}, nil
	}
}

func (srv *DiscoverService) save(ad *Ad) {
	hexID := ad.Identity.String()
	srv.cache[hexID] = ad
	srv.clean()
}

func (srv *DiscoverService) clean() {
	for hexID, p := range srv.cache {
		if time.Since(p.Timestamp) >= PresenceTimeout {
			delete(srv.cache, hexID)
		}
	}
}
