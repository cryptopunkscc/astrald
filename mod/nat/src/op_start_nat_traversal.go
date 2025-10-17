package nat

import (
	"bytes"
	mrand "math/rand"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

type opStartNatTraversal struct {
	// Active side fields
	Target string `query:"optional"` // if not empty act as initiator
	Out    string `query:"optional"`
}

func (mod *Module) OpStartNatTraversal(ctx *astral.Context, q shell.Query, args opStartNatTraversal) error {
	ch := astral.NewChannelFmt(q.Accept(), args.Out, args.Out)
	defer ch.Close()

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

		p, err := newConePuncher()
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		lp, err := p.Open(ctx)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		defer p.Close()

		routedQuery := query.New(ctx.Identity(), target, nat.MethodStartNatTraversal, &opStartNatTraversal{
			Out: args.Out,
		})

		routedQueryCtx := ctx.IncludeZone(astral.ZoneNetwork)
		peerCh, err := query.RouteChan(routedQueryCtx, mod.node, routedQuery)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		defer peerCh.Close()

		err = peerCh.Write(&nat.NatSignal{
			Signal:  nat.NatSignalTypeOffer,
			Session: p.Session(),
			IP:      localIP,
			Port:    astral.Uint16(lp),
		})
		if err != nil {
			mod.log.Info("peerCh.Write offer error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		ansObj, err := peerCh.Read()
		if err != nil {
			mod.log.Info("peerCh.ReadPayload answer error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		answer, ok := ansObj.(*nat.NatSignal)
		if !ok || answer == nil || answer.Signal != nat.NatSignalTypeAnswer {
			mod.log.Info("invalid answer: %v", answer)
			return ch.Write(astral.NewError("invalid answer"))
		}

		if !bytes.Equal(answer.Session, p.Session()) {
			mod.log.Info("session mismatch in answer: %v", answer.Session)
			return ch.Write(astral.NewError("session mismatch in answer"))
		}

		peerIP := answer.IP
		peerPort := int(answer.Port)

		err = peerCh.Write(&nat.NatSignal{Signal: nat.NatSignalTypeReady,
			Session: p.Session()})
		if err != nil {
			mod.log.Info("peerCh.Write ready error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		goObj, err := peerCh.Read()
		if err != nil {
			mod.log.Info("peerCh.ReadPayload go error: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}

		goSig, ok := goObj.(*nat.NatSignal)
		if !ok || goSig == nil || goSig.Signal != nat.NatSignalTypeGo {
			mod.log.Info("invalid go signal: %v", goSig)
			return ch.Write(astral.NewError("invalid go signal"))
		}
		if !bytes.Equal(goSig.Session, p.Session()) {
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
			Signal:  nat.NatSignalTypeResult,
			Session: p.Session(),
			IP:      punchResult.RemoteIP,
			Port:    punchResult.RemotePort,
		})
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		resObj, err := peerCh.Read()
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		result, ok := resObj.(*nat.NatSignal)
		if !ok || result == nil || result.Signal != nat.NatSignalTypeResult {
			return ch.Write(astral.NewError("invalid result signal"))
		}

		if !bytes.Equal(result.Session, p.Session()) {
			return ch.Write(astral.NewError("session mismatch in result signal"))
		}

		selfObserved := result.IP
		selfObservedPort := result.Port
		peerObserved := punchResult.RemoteIP
		peerObservedPort := punchResult.RemotePort

		mod.log.Info("NAT Traversal punch success: peer at %v:%v us at %v:%v"+
			"", peerIP, peerPort, selfObserved, selfObservedPort)

		traversedPair := nat.EndpointPair{
			PeerA: nat.PeerEndpoint{
				Identity: ctx.Identity(),
				Endpoint: &utp.Endpoint{IP: selfObserved,
					Port: selfObservedPort},
			},
			PeerB: nat.PeerEndpoint{
				Identity: target,
				Endpoint: &utp.Endpoint{IP: peerObserved,
					Port: peerObservedPort},
			},
			CreatedAt: astral.Time(time.Now()),
		}

		err = mod.addTraversedPair(traversedPair)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		err = ch.Write(&traversedPair)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		return nil
	}

	// Responder logic
	localIP := ips[0]

	obj, err := ch.Read()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	offer, ok := obj.(*nat.NatSignal)
	if !ok || offer == nil || offer.Signal != nat.NatSignalTypeOffer {
		return ch.Write(astral.NewError("invalid offer"))
	}
	if len(offer.Session) == 0 {
		return ch.Write(astral.NewError("missing session in offer"))
	}

	session := offer.Session
	peerIP := offer.IP
	peerPort := int(offer.Port)

	p, err := newConePuncher()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	lp, err := p.Open(ctx)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}
	defer p.Close()

	err = ch.Write(&nat.NatSignal{
		Signal:  nat.NatSignalTypeAnswer,
		Session: session,
		IP:      localIP,
		Port:    astral.Uint16(lp),
	})
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	readyObj, err := ch.Read()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	ready, ok := readyObj.(*nat.NatSignal)
	if !ok || ready == nil || ready.Signal != nat.NatSignalTypeReady {
		return ch.Write(astral.NewError("invalid ready signal"))
	}
	if !bytes.Equal(ready.Session, session) {
		return ch.Write(astral.NewError("session mismatch in ready signal"))
	}

	err = ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeGo, Session: session})
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

	punchResult, err := p.HolePunch(ctx, peerIP, peerPort)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	resObj, err := ch.Read()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	result, ok := resObj.(*nat.NatSignal)
	if !ok || result == nil || result.Signal != nat.NatSignalTypeResult {
		return ch.Write(astral.NewError("invalid result signal"))
	}
	if !bytes.Equal(result.Session, session) {
		return ch.Write(astral.NewError("session mismatch in result signal"))
	}

	err = ch.Write(&nat.NatSignal{
		Signal:  nat.NatSignalTypeResult,
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

	traversedPair := nat.EndpointPair{
		PeerA: nat.PeerEndpoint{
			Identity: ctx.Identity(),
			Endpoint: &utp.Endpoint{
				IP:   selfObserved,
				Port: selfObservedPort,
			},
		},
		PeerB: nat.PeerEndpoint{
			Identity: q.Caller(),
			Endpoint: &utp.Endpoint{
				IP:   peerObserved,
				Port: peerObservedPort,
			},
		},
		CreatedAt: astral.Time(time.Now()),
	}
	err = mod.addTraversedPair(traversedPair)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return nil
}
