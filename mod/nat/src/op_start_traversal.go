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

type opStartTraversal struct {
	// Active side fields
	Target string `query:"optional"` // if not empty act as initiator
	Out    string `query:"optional"`
}

func (mod *Module) OpStartTraversal(ctx *astral.Context, q shell.Query, args opStartTraversal) error {
	mod.log.Info("Starting NAT traversal operation")
	ch := astral.NewChannelFmt(q.Accept(), args.Out, args.Out)
	defer ch.Close()

	ips := mod.IP.PublicIPCandidates()
	mod.log.Info("Retrieved public IP candidates: %v", ips)
	if len(ips) == 0 {
		mod.log.Info("No suitable IP addresses found")
		return ch.Write(astral.NewError("no suitable IP addresses found"))
	}

	if args.Target != "" {
		mod.log.Info("Acting as initiator with target: %s", args.Target)
		// Initiator logic
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			mod.log.Info("Failed to resolve target identity: %v", err)
			return q.RejectWithCode(4)
		}
		mod.log.Info("Resolved target identity: %v", target)

		localIP := ips[0]
		mod.log.Info("Using local IP: %v", localIP)

		p, err := newConePuncher()
		if err != nil {
			mod.log.Info("Failed to create cone puncher: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Created cone puncher with session: %x", p.Session())

		lp, err := p.Open()
		if err != nil {
			mod.log.Info("Failed to open cone puncher: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Opened cone puncher on port: %d", lp)

		defer p.Close()

		routedQuery := query.New(ctx.Identity(), target, nat.MethodStartNatTraversal, &opStartTraversal{
			Out: args.Out,
		})
		mod.log.Info("Created routed query to target")

		routedQueryCtx := ctx.IncludeZone(astral.ZoneNetwork)
		peerCh, err := query.RouteChan(routedQueryCtx, mod.node, routedQuery)
		if err != nil {
			mod.log.Info("Failed to route channel: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Routed channel established")

		defer peerCh.Close()

		err = peerCh.Write(&nat.NatSignal{
			Signal:  nat.NatSignalTypeOffer,
			Session: p.Session(),
			IP:      localIP,
			Port:    astral.Uint16(lp),
		})
		if err != nil {
			mod.log.Info("Failed to write offer: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Sent offer signal")

		ansObj, err := peerCh.Read()
		if err != nil {
			mod.log.Info("Failed to read answer: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Received answer object")

		answer, ok := ansObj.(*nat.NatSignal)
		if !ok || answer == nil || answer.Signal != nat.NatSignalTypeAnswer {
			mod.log.Info("Invalid answer signal: %v", answer)
			return ch.Write(astral.NewError("invalid answer"))
		}
		mod.log.Info("Answer signal is valid")

		if !bytes.Equal(answer.Session, p.Session()) {
			mod.log.Info("Session mismatch in answer: %x vs %x", answer.Session, p.Session())
			return ch.Write(astral.NewError("session mismatch in answer"))
		}
		mod.log.Info("Session matches in answer")

		peerIP := answer.IP
		peerPort := int(answer.Port)
		mod.log.Info("Peer IP and port: %v:%v", peerIP, peerPort)

		err = peerCh.Write(&nat.NatSignal{Signal: nat.NatSignalTypeReady,
			Session: p.Session()})
		if err != nil {
			mod.log.Info("Failed to write ready: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Sent ready signal")

		goObj, err := peerCh.Read()
		if err != nil {
			mod.log.Info("Failed to read go signal: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Received go object")

		goSig, ok := goObj.(*nat.NatSignal)
		if !ok || goSig == nil || goSig.Signal != nat.NatSignalTypeGo {
			mod.log.Info("Invalid go signal: %v", goSig)
			return ch.Write(astral.NewError("invalid go signal"))
		}
		mod.log.Info("Go signal is valid")

		if !bytes.Equal(goSig.Session, p.Session()) {
			mod.log.Info("Session mismatch in go signal: %x vs %x", goSig.Session, p.Session())
			return ch.Write(astral.NewError("session mismatch in go signal"))
		}
		mod.log.Info("Session matches in go signal")
		punchResult, err := p.HolePunch(ctx, peerIP, peerPort)
		if err != nil {
			mod.log.Info("Hole punch failed: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Hole punch succeeded: remote IP %v, port %d", punchResult.RemoteIP, punchResult.RemotePort)

		err = peerCh.Write(&nat.NatSignal{
			Signal:  nat.NatSignalTypeResult,
			Session: p.Session(),
			IP:      punchResult.RemoteIP,
			Port:    punchResult.RemotePort,
		})
		if err != nil {
			mod.log.Info("Failed to write result: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Sent result signal")

		resObj, err := peerCh.Read()
		if err != nil {
			mod.log.Info("Failed to read result: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Received result object")

		result, ok := resObj.(*nat.NatSignal)
		if !ok || result == nil || result.Signal != nat.NatSignalTypeResult {
			mod.log.Info("Invalid result signal: %v", result)
			return ch.Write(astral.NewError("invalid result signal"))
		}
		mod.log.Info("Result signal is valid")

		if !bytes.Equal(result.Session, p.Session()) {
			mod.log.Info("Session mismatch in result signal: %x vs %x", result.Session, p.Session())
			return ch.Write(astral.NewError("session mismatch in result signal"))
		}
		mod.log.Info("Session matches in result signal")

		selfObserved := result.IP
		selfObservedPort := result.Port
		peerObserved := punchResult.RemoteIP
		peerObservedPort := punchResult.RemotePort

		mod.log.Info("NAT Traversal punch success: peer at %v:%v us at %v:%v",
			peerIP, peerPort, selfObserved, selfObservedPort)

		traversedPair := nat.EndpointPair{
			PeerA: nat.PeerEndpoint{
				Identity: ctx.Identity(),
				Endpoint: utp.Endpoint{
					IP:   selfObserved,
					Port: selfObservedPort,
				},
			},
			PeerB: nat.PeerEndpoint{
				Identity: target,
				Endpoint: utp.Endpoint{
					IP:   peerObserved,
					Port: peerObservedPort,
				},
			},
			CreatedAt: astral.Time(time.Now()),
		}
		mod.log.Info("Created traversed pair: %+v", traversedPair)

		mod.addTraversedPair(traversedPair)

		mod.log.Info("Added traversed pair")

		err = ch.Write(&traversedPair)
		if err != nil {
			mod.log.Info("Failed to write traversed pair: %v", err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Info("Wrote traversed pair to channel")

		return nil
	}

	mod.log.Info("Acting as responder")
	// Responder logic
	localIP := ips[0]
	mod.log.Info("Using local IP: %v", localIP)

	obj, err := ch.Read()
	if err != nil {
		mod.log.Info("Failed to read offer: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}
	mod.log.Info("Received offer object")

	offer, ok := obj.(*nat.NatSignal)
	if !ok || offer == nil || offer.Signal != nat.NatSignalTypeOffer {
		mod.log.Info("Invalid offer signal: %v", offer)
		return ch.Write(astral.NewError("invalid offer"))
	}
	mod.log.Info("Offer signal is valid")

	if len(offer.Session) == 0 {
		mod.log.Info("Missing session in offer")
		return ch.Write(astral.NewError("missing session in offer"))
	}
	mod.log.Info("Session present in offer: %x", offer.Session)

	peerIP := offer.IP
	peerPort := int(offer.Port)
	mod.log.Info("Peer IP and port from offer: %v:%v", peerIP, peerPort)

	p, err := newConePuncherWithSession(offer.Session)
	if err != nil {
		mod.log.Info("Failed to create cone puncher: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}
	mod.log.Info("Created cone puncher with session: %x", p.Session())

	lp, err := p.Open()
	if err != nil {
		mod.log.Info("Failed to open cone puncher: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}
	mod.log.Info("Opened cone puncher on port: %d", lp)
	defer p.Close()

	err = ch.Write(&nat.NatSignal{
		Signal:  nat.NatSignalTypeAnswer,
		Session: p.Session(),
		IP:      localIP,
		Port:    astral.Uint16(lp),
	})
	if err != nil {
		mod.log.Info("Failed to write answer: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}
	mod.log.Info("Sent answer signal")

	readyObj, err := ch.Read()
	if err != nil {
		mod.log.Info("Failed to read ready: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}
	mod.log.Info("Received ready object")

	ready, ok := readyObj.(*nat.NatSignal)
	if !ok || ready == nil || ready.Signal != nat.NatSignalTypeReady {
		mod.log.Info("Invalid ready signal: %v", ready)
		return ch.Write(astral.NewError("invalid ready signal"))
	}
	mod.log.Info("Ready signal is valid")

	if !bytes.Equal(ready.Session, p.Session()) {
		mod.log.Info("Session mismatch in ready signal: %x vs %x", ready.Session, p.Session())
		return ch.Write(astral.NewError("session mismatch in ready signal"))
	}
	mod.log.Info("Session matches in ready signal")

	err = ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeGo, Session: p.Session()})
	if err != nil {
		mod.log.Info("Failed to write go: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}
	mod.log.Info("Sent go signal")

	time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)
	mod.log.Info("Slept for random duration")

	punchResult, err := p.HolePunch(ctx, peerIP, peerPort)
	if err != nil {
		mod.log.Info("Hole punch failed: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}
	mod.log.Info("Hole punch succeeded: remote IP %v, port %d", punchResult.RemoteIP, punchResult.RemotePort)

	resObj, err := ch.Read()
	if err != nil {
		mod.log.Info("Failed to read result: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}
	mod.log.Info("Received result object")

	result, ok := resObj.(*nat.NatSignal)
	if !ok || result == nil || result.Signal != nat.NatSignalTypeResult {
		mod.log.Info("Invalid result signal: %v", result)
		return ch.Write(astral.NewError("invalid result signal"))
	}
	mod.log.Info("Result signal is valid")

	if !bytes.Equal(result.Session, p.Session()) {
		mod.log.Info("Session mismatch in result signal: %x vs %x", result.Session, p.Session())
		return ch.Write(astral.NewError("session mismatch in result signal"))
	}
	mod.log.Info("Session matches in result signal")

	err = ch.Write(&nat.NatSignal{
		Signal:  nat.NatSignalTypeResult,
		Session: p.Session(),
		IP:      punchResult.RemoteIP,   // FIX: use observed IP
		Port:    punchResult.RemotePort, // FIX: use observed Port
	})
	if err != nil {
		mod.log.Info("Failed to write result: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}
	mod.log.Info("Sent result signal")

	selfObserved := result.IP
	selfObservedPort := result.Port
	peerObserved := punchResult.RemoteIP
	peerObservedPort := punchResult.RemotePort

	traversedPair := nat.EndpointPair{
		PeerA: nat.PeerEndpoint{
			Identity: ctx.Identity(),
			Endpoint: utp.Endpoint{
				IP:   selfObserved,
				Port: selfObservedPort,
			},
		},
		PeerB: nat.PeerEndpoint{
			Identity: q.Caller(),
			Endpoint: utp.Endpoint{
				IP:   peerObserved,
				Port: peerObservedPort,
			},
		},
		CreatedAt: astral.Time(time.Now()),
	}
	mod.log.Info("Created traversed pair: %+v", traversedPair)
	mod.addTraversedPair(traversedPair)

	mod.log.Info("Added traversed pair")

	return nil
}
