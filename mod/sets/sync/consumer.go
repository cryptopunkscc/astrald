package sync

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/arl"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"time"
)

type Consumer struct {
	router net.Router
	arl    *arl.ARL
}

func NewConsumer(arl *arl.ARL, router net.Router) *Consumer {
	return &Consumer{router: router, arl: arl}
}

func (c *Consumer) Sync(ctx context.Context, since time.Time) (diff Diff, err error) {
	var params = router.Params{}

	if !since.IsZero() {
		params.SetUnixNano("since", since)
	}

	var query = net.NewQuery(
		c.arl.Caller,
		c.arl.Target,
		router.Query(c.arl.Query, params),
	)

	conn, err := net.Route(ctx, c.router, query)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var op byte
		err = cslq.Decode(conn, "c", &op)
		if err != nil {
			return
		}

		switch op {
		case opDone: // done
			var timestamp int64
			err = cslq.Decode(conn, "q", &timestamp)
			diff.Time = time.Unix(0, timestamp)
			return

		case opAdd: // add
			var dataID data.ID
			err = cslq.Decode(conn, "v", &dataID)
			if err != nil {
				return
			}

			diff.Updates = append(diff.Updates, Update{
				DataID:  dataID,
				Present: true,
			})

		case opRemove: // remove
			var dataID data.ID
			err = cslq.Decode(conn, "v", &dataID)
			if err != nil {
				return
			}

			diff.Updates = append(diff.Updates, Update{
				DataID:  dataID,
				Present: false,
			})

		case opResync:
			return diff, ErrResyncRequired

		default:
			return diff, ErrProtocolError
		}
	}
}
