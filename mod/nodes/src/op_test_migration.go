package nodes

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

// Args: only Target; if empty, act as responder.
type opTestMigrationArgs struct {
	Target string
}

// OpTestMigration: single-op sender/receiver.
// - If Target is provided: initiator. Connects to peer's nodes.test_migration and sends "Hello World" every 5s.
// - If Target is empty: responder. Accepts messages and logs them.
func (mod *Module) OpTestMigration(ctx *astral.Context, q shell.Query, args opTestMigrationArgs) error {
	// Initiator branch: Target provided
	if args.Target != "" {
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			return q.RejectWithCode(4)
		}

		ch := astral.NewChannel(q.Accept())
		defer ch.Close()

		peerQuery := query.New(ctx.Identity(), target, "nodes.test_migration", &opTestMigrationArgs{})
		peerCh, err := query.RouteChan(
			ctx.IncludeZone(astral.ZoneNetwork),
			mod.node,
			peerQuery,
		)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		defer peerCh.Close()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		msg := peerQuery.Nonce

		// Send immediately, then on ticks
		if err := peerCh.Write(&msg); err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		for {
			select {
			case <-ctx.Done():
				return ch.Write(&astral.Ack{})
			case <-ticker.C:
				if err := peerCh.Write(&msg); err != nil {
					return ch.Write(astral.NewError(err.Error()))
				}
			}
		}
	}

	// Responder branch: no Target provided
	ch := astral.NewChannel(q.Accept())
	defer ch.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			obj, err := ch.Read()
			if err != nil {
				return nil
			}

			if s, ok := obj.(*astral.Nonce); ok {
				session, ok := mod.peers.sessions.Get(*s)
				if ok {
					mod.log.Logv(0,
						"[test_migration] from %v: session network",
						q.Caller(), session.stream.Network())
				}
				mod.log.Logv(0, "[test_migration] from %v: %s", q.Caller(), s.String())
			}
		}
	}
}
