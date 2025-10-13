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

// initiateNatTraversal runs the active side of NAT traversal coordination.
func (mod *Module) initiateNatTraversal(ctx *astral.Context, q shell.Query, args opStartNatTraversal, ch *astral.Channel) error {
	target, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return q.RejectWithCode(4)
	}

	ips := mod.IP.FindIPCandidates()
	if len(ips) == 0 {
		return fmt.Errorf("no IP candidates available")
	}
	localIP := ips[0]

	session := make([]byte, 16)
	_, err = rand.Read(session)
	if err != nil {
		return err
	}

	p := newConePuncher(session)
	lp, err := p.Open(ctx)
	if err != nil {
		return err
	}

	routedQuery := query.New(ctx.Identity(), target, nat.MethodStartNatTraversal, &opStartNatTraversal{})
	peerCh, err := query.RouteChan(ctx, mod.node, routedQuery)
	if err != nil {
		return err
	}
	defer peerCh.Close()

	signal := &nat.NatSignal{
		Type:    nat.NatSignalTypeOffer,
		Session: session,
		IP:      localIP,
		Port:    astral.Uint16(lp),
	}
	err = peerCh.Write(signal)
	if err != nil {
		return err
	}

	ansObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}
	answer, _ := ansObj.(*nat.NatSignal)
	if answer == nil || answer.Type != nat.NatSignalTypeAnswer {
		return errors.New("invalid answer")
	}

	peerIP := answer.IP
	peerPort := int(answer.Port)

	ready := nat.NatSignal{Type: astral.String(nat.NatSignalTypeReady)}
	err = peerCh.Write(&ready)
	if err != nil {
		return err
	}

	goObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}
	goSig, _ := goObj.(*nat.NatSignal)
	if goSig == nil || goSig.Type != nat.NatSignalTypeGo {
		return fmt.Errorf("invalid go signal")
	}

	time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

	punchResult, err := p.HolePunch(ctx, peerIP, peerPort)
	if err != nil {
		mod.log.Error("hole punch failed: %v", err)
		return err
	}

	resultSignal := &nat.NatSignal{
		Type: astral.String(nat.NatSignalTypeResult),
		IP:   peerIP,
		Port: punchResult.RemotePort,
	}
	err = peerCh.Write(resultSignal)
	if err != nil {
		return err
	}

	resObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}
	result, ok := resObj.(*nat.NatSignal)
	if !ok || result.Type != nat.NatSignalTypeResult {
		return fmt.Errorf("invalid result signal")
	}

	selfObserved := result.IP
	selfObservedPort := result.Port
	peerObserved := punchResult.RemoteIP
	peerObservedPort := punchResult.RemotePort

	mod.log.Info("NAT traversal success:")
	mod.log.Info("Our external address as seen by peer: %v:%v",
		selfObserved, selfObservedPort)
	mod.log.Info("Peer external address as seen by us: %v:%v",
		peerObserved, peerObservedPort)

	ch.Write(&nat.TraversalResult{
		PeerObservedIP:   peerObserved,
		PeerObservedPort: peerObservedPort,
		ObservedIP:       selfObserved,
		ObservedPort:     selfObservedPort,
	})
	return nil
}

// respondNatTraversal runs the passive side of NAT traversal coordination.
func (mod *Module) respondNatTraversal(ctx *astral.Context, args opStartNatTraversal, ch *astral.Channel) error {
	ips := mod.IP.FindIPCandidates()
	if len(ips) == 0 {
		return fmt.Errorf("no IP candidates available")
	}
	localIP := ips[0]

	obj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}
	offer, _ := obj.(*nat.NatSignal)
	if offer == nil || offer.Type != nat.NatSignalTypeOffer {
		return errors.New("invalid offer")
	}

	session := offer.Session
	peerIP := offer.IP
	peerPort := int(offer.Port)

	p := newConePuncher(session)
	lp, err := p.Open(ctx)
	if err != nil {
		return err
	}

	answer := &nat.NatSignal{
		Type:    nat.NatSignalTypeAnswer,
		Session: session,
		IP:      localIP,
		Port:    astral.Uint16(lp),
	}
	err = ch.Write(answer)
	if err != nil {
		return err
	}

	readyObj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}
	ready, _ := readyObj.(*nat.NatSignal)
	if ready == nil || ready.Type != nat.NatSignalTypeReady {
		return fmt.Errorf("invalid ready signal")
	}

	goSig := nat.NatSignal{Type: astral.String(nat.NatSignalTypeGo)}
	err = ch.Write(&goSig)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

	punchResult, err := p.HolePunch(ctx, peerIP, peerPort)
	if err != nil {
		mod.log.Error("hole punch failed: %v", err)
		return err
	}

	resObj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}
	result, ok := resObj.(*nat.NatSignal)
	if !ok || result.Type != nat.NatSignalTypeResult {
		return fmt.Errorf("invalid result signal")
	}

	response := &nat.NatSignal{
		Type: nat.NatSignalTypeResult,
		IP:   peerIP,
		Port: punchResult.RemotePort,
	}
	err = ch.Write(response)
	if err != nil {
		return err
	}

	mod.log.Info("NAT traversal result sent: observed peer at %s:%d", peerIP, int(punchResult.RemotePort))
	return nil
}
