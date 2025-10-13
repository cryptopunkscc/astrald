package nat

import (
	"crypto/rand"
	"errors"
	"fmt"
	mrand "math/rand"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opStartNatTraversal struct {
	// Active side fields
	Target string `query:"optional"` // if not empty act as initiator
	//
	Out string `query:"optional"`
}

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

	// generate random session id
	session := make([]byte, 16)
	if _, err := rand.Read(session); err != nil {
		return err
	}

	// prepare puncher and open UDP socket (puncher owns it)
	p := newConePuncher(session)
	lp, err := p.Open(ctx)
	if err != nil {
		return err
	}

	// Call peer with the same method
	routedQuery := query.New(ctx.Identity(), target, nat.MethodStartNatTraversal, &opStartNatTraversal{})

	// Route to passive side
	peerCh, err := query.RouteChan(ctx, mod.node, routedQuery)
	if err != nil {
		return err
	}
	defer peerCh.Close()

	if err := peerCh.Write(&nat.NatSignal{
		Type:    nat.NatSignalTypeOffer,
		Session: session,
		IP:      ips[0],
		Port:    astral.Uint16(lp),
	}); err != nil {
		return err
	}

	// wait for answer
	ansObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}

	answer, _ := ansObj.(*nat.NatSignal)
	if answer == nil || string(answer.Type) != nat.NatSignalTypeAnswer {
		return errors.New("invalid answer")
	}
	peerIP := answer.IP
	peerPort := int(answer.Port)

	// send ready
	ready := nat.NatSignal{Type: astral.String(nat.NatSignalTypeReady)}
	if err := peerCh.Write(&ready); err != nil {
		return err
	}

	// wait for go
	natSignalObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}

	goSig, _ := natSignalObj.(*nat.NatSignal)
	if goSig == nil || string(goSig.Type) != nat.NatSignalTypeGo {
		return fmt.Errorf(`invalid go signal`)
	}

	// small random delay
	time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

	// start punching (reuse the already opened socket)
	if _, err := p.HolePunch(ctx, peerIP, peerPort); err != nil {
		mod.log.Error("hole punch failed: %v", err)
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

	// read offer
	obj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}

	offer, _ := obj.(*nat.NatSignal)
	if offer == nil || string(offer.Type) != nat.NatSignalTypeOffer {
		return errors.New("invalid offer")
	}

	session := offer.Session
	peerIP := offer.IP
	peerPort := int(offer.Port)

	// prepare puncher and open UDP socket (puncher owns it)
	p := newConePuncher(session)
	lp, err := p.Open(ctx)
	if err != nil {
		return err
	}

	err = ch.Write(&nat.NatSignal{
		Type:    nat.NatSignalTypeAnswer,
		Session: session,
		IP:      ips[0],
		Port:    astral.Uint16(lp),
	})
	if err != nil {
		return err
	}

	// wait for ready
	readyObj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}

	ready, _ := readyObj.(*nat.NatSignal)
	if ready == nil || string(ready.Type) != nat.NatSignalTypeReady {
		return fmt.Errorf(`invalid ready signal`)
	}

	// send go
	goSig := nat.NatSignal{Type: astral.String(nat.NatSignalTypeGo)}
	if err := ch.Write(&goSig); err != nil {
		return err
	}

	// small random delay
	time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

	// start punching (reuse the already opened socket)
	if _, err := p.HolePunch(ctx, peerIP, peerPort); err != nil {
		mod.log.Error("hole punch failed: %v", err)
		return err
	}

	return nil
}
