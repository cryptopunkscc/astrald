package nat

import (
	"crypto/rand"
	"errors"
	"fmt"
	mrand "math/rand"
	"net"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opStartNatTraversal struct {
	// Active side fields
	Target string `query:"optional"` // if not empty act as initiator
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

	// Bind UDP socket to get local port (close after signaling)
	udp, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return fmt.Errorf("udp listen: %w", err)
	}
	lp := 0
	if ua, ok := udp.LocalAddr().(*net.UDPAddr); ok {
		lp = ua.Port
	}
	udp.Close()

	// generate random session id
	session := make([]byte, 16)
	if _, err := rand.Read(session); err != nil {
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

	// send offer
	offer := nat.NatSignal{Type: astral.String("offer"), Session: astral.Bytes(session), IP: ip.IP(ips[0]), Port: astral.Uint16(lp)}
	if err := peerCh.Write(&offer); err != nil {
		return err
	}

	// wait for answer
	ansObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}
	answer, _ := ansObj.(*nat.NatSignal)
	if answer == nil || string(answer.Type) != "answer" {
		return errors.New("invalid answer")
	}
	peerIP := answer.IP
	peerPort := int(answer.Port)

	// send ready
	ready := nat.NatSignal{Type: astral.String("ready")}
	if err := peerCh.Write(&ready); err != nil {
		return err
	}

	// wait for go
	goObj, err := peerCh.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}
	goSig, _ := goObj.(*nat.NatSignal)
	if goSig == nil || string(goSig.Type) != "go" {
		return errors.New("invalid go")
	}

	// small random delay
	time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

	// start punching
	p := newConePuncher(session)
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
	if offer == nil || string(offer.Type) != "offer" {
		return errors.New("invalid offer")
	}
	session := []byte(offer.Session)
	peerIP := offer.IP
	peerPort := int(offer.Port)

	// Bind UDP socket to get local port (close after signaling)
	udp, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return fmt.Errorf("udp listen: %w", err)
	}
	lp := 0
	if ua, ok := udp.LocalAddr().(*net.UDPAddr); ok {
		lp = ua.Port
	}
	udp.Close()

	// send answer
	answer := nat.NatSignal{Type: astral.String("answer"), Session: astral.Bytes(session), IP: ip.IP(ips[0]), Port: astral.Uint16(lp)}
	if err := ch.Write(&answer); err != nil {
		return err
	}

	// wait for ready
	readyObj, err := ch.ReadPayload(nat.NatSignal{}.ObjectType())
	if err != nil {
		return err
	}
	ready, _ := readyObj.(*nat.NatSignal)
	if ready == nil || string(ready.Type) != "ready" {
		return errors.New("invalid ready")
	}

	// send go
	goSig := nat.NatSignal{Type: astral.String("go")}
	if err := ch.Write(&goSig); err != nil {
		return err
	}

	// small random delay
	time.Sleep(time.Duration(mrand.Intn(100)) * time.Millisecond)

	// start punching
	p := newConePuncher(session)
	if _, err := p.HolePunch(ctx, peerIP, peerPort); err != nil {
		mod.log.Error("hole punch failed: %v", err)
		return err
	}

	return nil
}
