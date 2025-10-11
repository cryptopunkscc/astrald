package nat

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

type opStartNatTraversal struct {
	// Active side fields
	Target  string `query:"optional"`
	Out     string `query:"optional"`
	Session []byte `query:"optional"`
}

// FIXME: adjust error handling to standard

func (mod *Module) OpStartNatTraversal(ctx *astral.Context, q shell.Query,
	args opStartNatTraversal) error {
	ips := mod.IP.FindIPCandidates()
	if len(ips) == 0 {
		return fmt.Errorf("no IP candidates available")
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	// If Target is provided, act as initiator (active side) and route this same method to the peer.
	if args.Target != "" {
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			return q.RejectWithCode(4)
		}

		session := make([]byte, 16)
		if _, err := rand.Read(session); err != nil {
			return err
		}

		// Call peer with the same method, passing the session
		queryArgs := &opStartNatTraversal{Session: session}
		routedQuery := query.New(ctx.Identity(), target, nat.MethodStartNatTraversal, queryArgs)

		passiveSideCh, err := query.RouteChan(ctx, mod.node, routedQuery)
		if err != nil {
			return err
		}
		defer passiveSideCh.Close()

		// receive target's endpoint
		targetObj, err := passiveSideCh.ReadPayload((&utp.Endpoint{}).ObjectType())
		if err != nil {
			return err
		}

		targetEp, _ := targetObj.(*utp.Endpoint)
		if targetEp == nil || targetEp.IsZero() {
			return errors.New("invalid target endpoint")
		}

		// send our endpoint
		initEndpoint := &utp.Endpoint{
			IP:   ips[0],
			Port: astral.Uint16(mod.UTP.ListenPort()),
		}
		err = passiveSideCh.Write(initEndpoint)
		if err != nil {
			return err
		}

		mod.log.Info("NAT traversal info exchanged: local=%v, remote=%v", initEndpoint, targetEp)
		// Start punching (active side)
		p := newConePuncher(int(targetEp.Port), session)
		if _, err := p.HolePunch(ctx, targetEp.IP); err != nil {
			mod.log.Error("hole punch failed: %v", err)
			return err
		}

		// FIXME: return pair <{IP, Port>
		return nil
	}

	// Passive side (responder) behavior: accept channel, exchange endpoints, start punching
	if len(args.Session) == 0 {
		return errors.New("session required for responder")
	}

	// send local endpoint
	localEndpoint := &utp.Endpoint{
		IP:   ips[0],
		Port: astral.Uint16(mod.UTP.ListenPort()),
	}
	if err := ch.Write(localEndpoint); err != nil {
		return err
	}

	// read initiator endpoint
	initObj, err := ch.ReadPayload((&utp.Endpoint{}).ObjectType())
	if err != nil {
		return err
	}
	initEp, _ := initObj.(*utp.Endpoint)
	if initEp == nil || initEp.IsZero() {
		return fmt.Errorf("invalid initiator endpoint")
	}

	mod.log.Info("NAT traversal info exchanged: local=%v, remote=%v", localEndpoint, initEp)

	// Start punching (passive side)
	p := newConePuncher(int(initEp.Port), args.Session)
	if _, err := p.HolePunch(ctx, initEp.IP); err != nil {
		mod.log.Error("hole punch failed: %v", err)
		return err
	}

	return nil
}
