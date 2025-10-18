package nat

import (
	"bytes"
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

		p, err := newConePuncher()
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		lp, err := p.Open()
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		defer p.Close()

		peerCh, err := query.RouteChan(ctx.IncludeZone(astral.ZoneNetwork), mod.node, query.New(ctx.Identity(), target, nat.MethodStartNatTraversal, &opStartTraversal{
			Out: args.Out,
		}))
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		defer peerCh.Close()

		err = peerCh.Write(&nat.NatSignal{
			Signal:  nat.NatSignalTypeOffer,
			Session: p.Session(),
			IP:      ips[0],
			Port:    astral.Uint16(lp),
		})
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		ansObj, err := peerCh.Read()
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		answer, ok := ansObj.(*nat.NatSignal)
		if !ok || answer == nil || answer.Signal != nat.NatSignalTypeAnswer {
			return ch.Write(astral.NewError("invalid answer"))
		}

		if !bytes.Equal(answer.Session, p.Session()) {
			return ch.Write(astral.NewError("session mismatch in answer"))
		}

		err = peerCh.Write(&nat.NatSignal{Signal: nat.NatSignalTypeReady,
			Session: p.Session()})
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		goObj, err := peerCh.Read()
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		goSig, ok := goObj.(*nat.NatSignal)
		if !ok || goSig == nil || goSig.Signal != nat.NatSignalTypeGo {
			return ch.Write(astral.NewError("invalid go signal"))
		}

		if !bytes.Equal(goSig.Session, p.Session()) {
			return ch.Write(astral.NewError("session mismatch in go signal"))
		}

		punchResult, err := p.HolePunch(ctx, answer.IP, int(answer.Port))
		if err != nil {
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

		traversedPair := nat.EndpointPair{
			PeerA: nat.PeerEndpoint{
				Identity: ctx.Identity(),
				Endpoint: utp.Endpoint{
					IP:   result.IP,
					Port: result.Port,
				},
			},
			PeerB: nat.PeerEndpoint{
				Identity: target,
				Endpoint: utp.Endpoint{
					IP:   punchResult.RemoteIP,
					Port: punchResult.RemotePort,
				},
			},
			CreatedAt: astral.Time(time.Now()),
		}

		mod.addTraversedPair(traversedPair)

		err = ch.Write(&traversedPair)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		return nil
	}

	// Responder logic

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

	p, err := newConePuncherWithSession(offer.Session)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	lp, err := p.Open()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}
	defer p.Close()

	err = ch.Write(&nat.NatSignal{
		Signal:  nat.NatSignalTypeAnswer,
		Session: p.Session(),
		IP:      ips[0],
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

	if !bytes.Equal(ready.Session, p.Session()) {
		return ch.Write(astral.NewError("session mismatch in ready signal"))
	}

	err = ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeGo, Session: p.Session()})
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	punchResult, err := p.HolePunch(ctx, offer.IP, int(offer.Port))
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

	if !bytes.Equal(result.Session, p.Session()) {
		return ch.Write(astral.NewError("session mismatch in result signal"))
	}

	err = ch.Write(&nat.NatSignal{
		Signal:  nat.NatSignalTypeResult,
		Session: p.Session(),
		IP:      punchResult.RemoteIP,
		Port:    punchResult.RemotePort,
	})
	if err != nil {
		mod.log.Info("Failed to write result: %v", err)
		return ch.Write(astral.NewError(err.Error()))
	}

	traversedPair := nat.EndpointPair{
		PeerA: nat.PeerEndpoint{
			Identity: ctx.Identity(),
			Endpoint: utp.Endpoint{
				IP:   result.IP,
				Port: result.Port,
			},
		},
		PeerB: nat.PeerEndpoint{
			Identity: q.Caller(),
			Endpoint: utp.Endpoint{
				IP:   punchResult.RemoteIP,
				Port: punchResult.RemotePort,
			},
		},
		CreatedAt: astral.Time(time.Now()),
	}

	mod.addTraversedPair(traversedPair)

	return nil
}
