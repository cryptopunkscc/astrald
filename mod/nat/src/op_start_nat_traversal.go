package nat

import (
	"bytes"
	"crypto/rand"
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
	Out    string `query:"optional"`
}

func (mod *Module) OpStartNatTraversal(ctx *astral.Context, q shell.Query, args opStartNatTraversal) error {

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer func() { _ = ch.Close() }()

	ips := mod.IP.PublicIPCandidates()
	if len(ips) == 0 {
		return ch.Write(astral.NewError("no suitable IP addresses found"))
	}

	if args.Target != "" {
		// Initiator logic
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			return q.RejectWithCode(4)
		}

		localIP := ips[0]

		session := make([]byte, 16)
		_, err = rand.Read(session)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		p := newConePuncher(session)
		lp, err := p.Open(ctx)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		defer func() { _ = p.Close() }()

		routedQuery := query.New(ctx.Identity(), target, nat.MethodStartNatTraversal, &opStartNatTraversal{})
		peerCh, err := query.RouteChan(ctx, mod.node, routedQuery)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		defer func() { _ = peerCh.Close() }()

		err = peerCh.Write(&nat.NatSignal{
			Type:    nat.NatSignalTypeOffer,
			Session: session,
			IP:      localIP,
			Port:    astral.Uint16(lp),
		})
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		ansObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		answer, ok := ansObj.(*nat.NatSignal)
		if !ok || answer == nil || answer.Type != astral.String(nat.NatSignalTypeAnswer) {
			return ch.Write(astral.NewError("invalid answer"))
		}
		if !bytes.Equal(answer.Session, session) {
			return ch.Write(astral.NewError("session mismatch in answer"))
		}

		peerIP := answer.IP
		peerPort := int(answer.Port)

		err = peerCh.Write(&nat.NatSignal{Type: astral.String(nat.NatSignalTypeReady), Session: session})
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		goObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		goSig, ok := goObj.(*nat.NatSignal)
		if !ok || goSig == nil || goSig.Type != astral.String(nat.NatSignalTypeGo) {
			return ch.Write(astral.NewError("invalid go signal"))
		}
		if !bytes.Equal(goSig.Session, session) {
			return ch.Write(astral.NewError("session mismatch in go signal"))
		}

		time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

		punchResult, err := p.HolePunch(ctx, peerIP, peerPort)
		if err != nil {
			mod.log.Error("hole punch failed: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		err = peerCh.Write(&nat.NatSignal{
			Type:    nat.NatSignalTypeResult,
			Session: session,
			IP:      punchResult.RemoteIP,   // FIX: use observed IP
			Port:    punchResult.RemotePort, // FIX: use observed Port
		})
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		resObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		result, ok := resObj.(*nat.NatSignal)
		if !ok || result == nil || result.Type != nat.NatSignalTypeResult {
			return ch.Write(astral.NewError("invalid result signal"))
		}
		if !bytes.Equal(result.Session, session) {
			return ch.Write(astral.NewError("session mismatch in result signal"))
		}

		selfObserved := result.IP
		selfObservedPort := result.Port
		peerObserved := punchResult.RemoteIP
		peerObservedPort := punchResult.RemotePort

		mod.log.Info("NAT traversal success:")
		mod.log.Info("Our external address as seen by peer: %v:%v", selfObserved, selfObservedPort)
		mod.log.Info("Peer external address as seen by us: %v:%v", peerObserved, peerObservedPort)

		err = ch.Write(&nat.TraversalResult{
			PeerObservedIP:   peerObserved,
			PeerObservedPort: peerObservedPort,
			ObservedIP:       selfObserved,
			ObservedPort:     selfObservedPort,
		})
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		return nil
	}

	// Responder logic
	localIP := ips[0]

	obj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	offer, ok := obj.(*nat.NatSignal)
	if !ok || offer == nil || offer.Type != nat.NatSignalTypeOffer {
		return ch.Write(astral.NewError("invalid offer"))
	}
	if len(offer.Session) == 0 {
		return ch.Write(astral.NewError("missing session in offer"))
	}

	session := offer.Session
	peerIP := offer.IP
	peerPort := int(offer.Port)

	p := newConePuncher(session)
	lp, err := p.Open(ctx)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}
	defer func() { _ = p.Close() }()

	err = ch.Write(&nat.NatSignal{
		Type:    nat.NatSignalTypeAnswer,
		Session: session,
		IP:      localIP,
		Port:    astral.Uint16(lp),
	})
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	readyObj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	ready, ok := readyObj.(*nat.NatSignal)
	if !ok || ready == nil || ready.Type != nat.NatSignalTypeReady {
		return ch.Write(astral.NewError("invalid ready signal"))
	}
	if !bytes.Equal(ready.Session, session) {
		return ch.Write(astral.NewError("session mismatch in ready signal"))
	}

	err = ch.Write(&nat.NatSignal{Type: nat.NatSignalTypeGo, Session: session})
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

	punchResult, err := p.HolePunch(ctx, peerIP, peerPort)
	if err != nil {
		mod.log.Error("hole punch failed: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}

	resObj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	result, ok := resObj.(*nat.NatSignal)
	if !ok || result == nil || result.Type != nat.NatSignalTypeResult {
		return ch.Write(astral.NewError("invalid result signal"))
	}
	if !bytes.Equal(result.Session, session) {
		return ch.Write(astral.NewError("session mismatch in result signal"))
	}

	err = ch.Write(&nat.NatSignal{
		Type:    nat.NatSignalTypeResult,
		Session: session,
		IP:      punchResult.RemoteIP,   // FIX: use observed IP
		Port:    punchResult.RemotePort, // FIX: use observed Port
	})
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	selfObserved := result.IP
	selfObservedPort := result.Port
	peerObserved := punchResult.RemoteIP
	peerObservedPort := punchResult.RemotePort

	mod.log.Info("NAT traversal result sent: observed peer at %s:%d", peerIP, int(punchResult.RemotePort))

	err = ch.Write(&nat.TraversalResult{
		PeerObservedIP:   peerObserved,
		PeerObservedPort: peerObservedPort,
		ObservedIP:       selfObserved,
		ObservedPort:     selfObservedPort,
	})
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return nil
}
