package nat

// NOTE: might  move to mod/nat
import (
	"crypto/rand"
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

type opNewNatTraversal struct {
	Target string
	Out    string `query:"optional"`
}

func (mod *Module) OpNewNatTraversal(ctx *astral.Context, q shell.Query,
	args opNewNatTraversal) (err error) {

	ips := mod.IP.FindIPCandidates()
	if len(ips) == 0 {
		return errors.New("no IP candidates available")
	}

	target, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return q.RejectWithCode(4)
	}

	// generate a random session id for this traversal
	session := make([]byte, 16)
	if _, err := rand.Read(session); err != nil {
		return err
	}

	queryArgs := &opStartNatTraversal{Session: session}

	routedQuery := query.New(ctx.Identity(), target,
		nat.MethodStartNatTraversal,
		queryArgs)

	// route and get a bidirectional channel for payload exchange
	ch, err := query.RouteChan(ctx, mod.node, routedQuery)
	if err != nil {
		return err
	}
	defer ch.Close()

	// receive target's endpoint
	targetObj, err := ch.ReadPayload((&utp.Endpoint{}).ObjectType())
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
	if err := ch.Write(initEndpoint); err != nil {
		return err
	}

	mod.log.Info("NAT traversal info exchanged: local=%v, remote=%v", initEndpoint, targetEp)

	// Start hole punching towards target IP using the same session (simultaneous punching)
	p := newConePuncher(int(targetEp.Port), session)
	if _, err := p.HolePunch(ctx, targetEp.IP); err != nil {
		mod.log.Error("hole punch failed: %v", err)
		return err
	}

	// acknowledge the shell query for UX completeness
	shellCh := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer shellCh.Close()

	return nil
}
