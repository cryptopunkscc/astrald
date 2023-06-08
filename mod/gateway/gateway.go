package gateway

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/gw"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/streams"
	"sync"
)

const queryConnect = "connect"

type Gateway struct {
	node node.Node
	log  *log.Logger
}

func (mod *Gateway) Run(ctx context.Context) error {
	var queries = services.NewQueryChan(8)

	service, err := mod.node.Services().Register(ctx, mod.node.Identity(), gw.PortName, queries.Push)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(128)
	for i := 0; i < 128; i++ {
		go func() {
			for query := range queries {
				mod.handleQuery(ctx, query)
			}
			wg.Done()
		}()
	}

	<-service.Done()
	close(queries)
	wg.Wait()

	return nil
}

func (mod *Gateway) handleQuery(ctx context.Context, query *services.Query) error {
	conn, err := query.Accept()
	if err != nil {
		return err
	}

	var c = cslq.NewEndec(conn)
	var cookie string

	err = c.Decode("[c]c", &cookie)
	if err != nil {
		return err
	}

	nodeID, err := id.ParsePublicKeyHex(cookie)
	if err != nil {
		return err
	}

	peer := mod.node.Network().Peers().Find(nodeID)
	if peer == nil {
		return errors.New("node unavailable")
	}

	lnk := peer.PreferredLink()
	if lnk == nil {
		return errors.New("node unavailable")
	}

	// check link health
	lnk.Health().Check()

	out, err := lnk.Query(ctx, queryConnect)
	if err != nil {
		conn.Close()
		return err
	}

	c.Encode("c", true)

	l, r, err := streams.Join(conn, out)

	mod.log.Logv(1, "conn for %s done (bytes read %d written %d)", peer.Identity(), l, r)

	return err
}
