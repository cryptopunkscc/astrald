package sync

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/arl"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type Consumer struct {
	router astral.Router
	arl    *arl.ARL
}

func NewConsumer(arl *arl.ARL, router astral.Router) *Consumer {
	return &Consumer{router: router, arl: arl}
}

func (c *Consumer) Sync(ctx context.Context, since time.Time) (diff Diff, err error) {
	var params = core.Params{}

	if !since.IsZero() {
		params.SetUnixNano("since", since)
	}

	var query = astral.NewQuery(
		c.arl.Caller,
		c.arl.Target,
		core.Query(c.arl.Query, params),
	)

	conn, err := astral.Route(ctx, c.router, query)
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
			var objectID object.ID
			err = cslq.Decode(conn, "v", &objectID)
			if err != nil {
				return
			}

			diff.Updates = append(diff.Updates, Update{
				ObjectID: objectID,
				Present:  true,
			})

		case opRemove: // remove
			var objectID object.ID
			err = cslq.Decode(conn, "v", &objectID)
			if err != nil {
				return
			}

			diff.Updates = append(diff.Updates, Update{
				ObjectID: objectID,
				Present:  false,
			})

		case opResync:
			return diff, ErrResyncRequired

		default:
			return diff, ErrProtocolError
		}
	}
}
