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
	In     string `query:"optional"`
}

func (mod *Module) OpStartNatTraversal(ctx *astral.Context, q shell.Query, args opStartNatTraversal) error {

	if args.In == "" {
		args.In = args.Out
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer func() { _ = ch.Close() }()

	ips := mod.IP.PublicIPCandidates()
	if len(ips) == 0 {
		mod.log.Info("no suitable IP addresses found: %v", "no suitable IP addresses found")
		return ch.Write(astral.NewError("no suitable IP addresses found"))
	}

	if args.Target != "" {
		// Initiator logic
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			mod.log.Info("ResolveIdentity error: %v", err)
			return q.RejectWithCode(4)
		}

		localIP := ips[0]

		session := make([]byte, 16)
		_, err = rand.Read(session)
		if err != nil {
			mod.log.Info("rand.Read error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		p := newConePuncher(session)
		lp, err := p.Open(ctx)
		if err != nil {
			mod.log.Info("p.Open error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		defer func() { _ = p.Close() }()

		routedQuery := query.New(ctx.Identity(), target, nat.MethodStartNatTraversal, &opStartNatTraversal{})

		peerCh, err := query.RouteChan(
			ctx.IncludeZone(astral.ZoneNetwork),
			mod.node,
			routedQuery,
		)
		if err != nil {
			mod.log.Info("RouteChan error: %v", err)
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
			mod.log.Info("peerCh.Write offer error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		ansObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
		if err != nil {
			mod.log.Info("peerCh.ReadPayload answer error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		answer, ok := ansObj.(*nat.NatSignal)
		if !ok || answer == nil || answer.Type != nat.NatSignalTypeAnswer {
			mod.log.Info("invalid answer: %v", answer)
			return ch.Write(astral.NewError("invalid answer"))
		}
		if !bytes.Equal(answer.Session, session) {
			mod.log.Info("session mismatch in answer: %v", answer.Session)
			return ch.Write(astral.NewError("session mismatch in answer"))
		}

		peerIP := answer.IP
		peerPort := int(answer.Port)

		err = peerCh.Write(&nat.NatSignal{Type: nat.NatSignalTypeReady, Session: session})
		if err != nil {
			mod.log.Info("peerCh.Write ready error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		goObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
		if err != nil {
			mod.log.Info("peerCh.ReadPayload go error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		goSig, ok := goObj.(*nat.NatSignal)
		if !ok || goSig == nil || goSig.Type != nat.NatSignalTypeGo {
			mod.log.Info("invalid go signal: %v", goSig)
			return ch.Write(astral.NewError("invalid go signal"))
		}
		if !bytes.Equal(goSig.Session, session) {
			mod.log.Info("session mismatch in go signal: %v", goSig.Session)
			return ch.Write(astral.NewError("session mismatch in go signal"))
		}

		time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

		punchResult, err := p.HolePunch(ctx, peerIP, peerPort)
		if err != nil {
			mod.log.Info("hole punch failed: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		err = peerCh.Write(&nat.NatSignal{
			Type:    nat.NatSignalTypeResult,
			Session: session,
			IP:      punchResult.RemoteIP,   // FIX: use observed IP
			Port:    punchResult.RemotePort, // FIX: use observed Port
		})
		if err != nil {
			mod.log.Info("peerCh.Write result error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		resObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
		if err != nil {
			mod.log.Info("peerCh.ReadPayload result error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		result, ok := resObj.(*nat.NatSignal)
		if !ok || result == nil || result.Type != nat.NatSignalTypeResult {
			mod.log.Info("invalid result signal: %v", result)
			return ch.Write(astral.NewError("invalid result signal"))
		}
		if !bytes.Equal(result.Session, session) {
			mod.log.Info("session mismatch in result signal: %v", result.Session)
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
			mod.log.Info("ch.Write traversal result error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		return nil
	}

	// Responder logic
	localIP := ips[0]

	obj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		mod.log.Info("ch.ReadPayload offer error: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}

	offer, ok := obj.(*nat.NatSignal)
	if !ok || offer == nil || offer.Type != nat.NatSignalTypeOffer {
		mod.log.Info("invalid offer: %v", offer)
		return ch.Write(astral.NewError("invalid offer"))
	}
	if len(offer.Session) == 0 {
		mod.log.Info("missing session in offer: %v", offer)
		return ch.Write(astral.NewError("missing session in offer"))
	}

	session := offer.Session
	peerIP := offer.IP
	peerPort := int(offer.Port)

	p := newConePuncher(session)
	lp, err := p.Open(ctx)
	if err != nil {
		mod.log.Info("p.Open error: %v", err)
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
		mod.log.Info("ch.Write answer error: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}

	readyObj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		mod.log.Info("ch.ReadPayload ready error: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}

	ready, ok := readyObj.(*nat.NatSignal)
	if !ok || ready == nil || ready.Type != nat.NatSignalTypeReady {
		mod.log.Info("invalid ready signal: %v", ready)
		return ch.Write(astral.NewError("invalid ready signal"))
	}
	if !bytes.Equal(ready.Session, session) {
		mod.log.Info("session mismatch in ready signal: %v", ready.Session)
		return ch.Write(astral.NewError("session mismatch in ready signal"))
	}

	err = ch.Write(&nat.NatSignal{Type: nat.NatSignalTypeGo, Session: session})
	if err != nil {
		mod.log.Info("ch.Write go error: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}

	time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

	punchResult, err := p.HolePunch(ctx, peerIP, peerPort)
	if err != nil {
		mod.log.Info("hole punch failed: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}

	resObj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		mod.log.Info("ch.ReadPayload result error: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}

	result, ok := resObj.(*nat.NatSignal)
	if !ok || result == nil || result.Type != nat.NatSignalTypeResult {
		mod.log.Info("invalid result signal: %v", result)
		return ch.Write(astral.NewError("invalid result signal"))
	}
	if !bytes.Equal(result.Session, session) {
		mod.log.Info("session mismatch in result signal: %v", result.Session)
		return ch.Write(astral.NewError("session mismatch in result signal"))
	}

	err = ch.Write(&nat.NatSignal{
		Type:    nat.NatSignalTypeResult,
		Session: session,
		IP:      punchResult.RemoteIP,   // FIX: use observed IP
		Port:    punchResult.RemotePort, // FIX: use observed Port
	})
	if err != nil {
		mod.log.Info("ch.Write result error: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}

	selfObserved := result.IP
	selfObservedPort := result.Port
	peerObserved := punchResult.RemoteIP
	peerObservedPort := punchResult.RemotePort

	mod.log.Info("NAT traversal result sent: observed peer at %v:%v", peerIP, int(punchResult.RemotePort))

	err = ch.Write(&nat.TraversalResult{
		PeerObservedIP:   peerObserved,
		PeerObservedPort: peerObservedPort,
		ObservedIP:       selfObserved,
		ObservedPort:     selfObservedPort,
	})
	if err != nil {
		mod.log.Info("ch.Write traversal result error: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}

	return nil
}
