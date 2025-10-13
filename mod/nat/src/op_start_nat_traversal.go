package nat

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

type opStartNatTraversal struct {
	// Active side fields
	Target  string `query:"optional"` // if not empty act as initiator
	Session []byte `query:"optional"` // empty only for active side at first
	//
	Out string `query:"optional"`
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

	if args.Target != "" {
		return mod.initiateNatTraversal(ctx, q, args, ch)
	}

	return mod.respondNatTraversal(ctx, args, ch)
}

func (mod *Module) initiateNatTraversal(ctx *astral.Context, q shell.Query,
	args opStartNatTraversal, ch *astral.Channel) error {
	target, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return q.RejectWithCode(4)
	}

	ips := mod.IP.FindIPCandidates()
	if len(ips) == 0 {
		return fmt.Errorf("no IP candidates available")
	}

	// generate random session id for this traversal (used to verify packets belong to this traversal)
	session := make([]byte, 16)
	_, err = rand.Read(session)
	if err != nil {
		return err
	}

	// Call peer with the same method, passing the session
	queryArgs := &opStartNatTraversal{Session: session}
	routedQuery := query.New(ctx.Identity(), target, nat.MethodStartNatTraversal, queryArgs)

	// Route to passive side
	passiveSideCh, err := query.RouteChan(ctx, mod.node, routedQuery)
	if err != nil {
		return err
	}
	defer passiveSideCh.Close()

	// send our endpoint
	initEndpoint := &utp.Endpoint{
		IP:   ips[0],
		Port: astral.Uint16(mod.UTP.ListenPort()),
	}
	err = passiveSideCh.Write(initEndpoint)
	if err != nil {
		return err
	}

	return nil
}

func (mod *Module) respondNatTraversal(ctx *astral.Context,
	args opStartNatTraversal, ch *astral.Channel) error {

	ips := mod.IP.FindIPCandidates()
	if len(ips) == 0 {
		return fmt.Errorf("no IP candidates available")
	}

	// Passive side (responder) behavior: accept channel, exchange endpoints, start punching
	if len(args.Session) == 0 {
		return errors.New("session required for responder")
	}

	// read initiator endpoint
	ipObj, err := ch.ReadPayload((&ip.IP{}).ObjectType())
	if err != nil {
		return err
	}

	initiatorIp, _ := ipObj.(*ip.IP)
	if initiatorIp == nil {
		return fmt.Errorf("invalid initiator endpoint")
	}

	err = ch.Write(&ips[0])
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	// Start punching (passive side)
	p := newConePuncher(args.Session)
	if _, err := p.HolePunch(ctx, *initiatorIp); err != nil {
		mod.log.Error("hole punch failed: %v", err)
		return err
	}

	return nil
}
