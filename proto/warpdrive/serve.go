package warpdrive

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/cslq/rpc"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"io"
	"log"
	"sync"
)

type Dispatcher struct {
	Service
	context.Context
	CallerId   string
	Authorized bool
	LogPrefix  string
	Api        wrapper.Api
	Conn       io.ReadWriteCloser
	Job        *sync.WaitGroup
	*cslq.Endec
	*log.Logger
}

func (d Dispatcher) Serve(
	dispatch func(d *Dispatcher) error,
) (err error) {
	d.Logger = NewLogger(d.LogPrefix)
	d.Endec = cslq.NewEndec(d.Conn)
	for err == nil {
		err = dispatch(&d)
		if err == nil {
			d.Println("OK")
		}
	}
	if errors.Is(err, errEnded) {
		err = nil
	}
	if err != nil {
		d.Println(Error(err, "Failed"))
	}
	return errors.Unwrap(err)
}

func nextCommand(d *Dispatcher) (cmd uint8, err error) {
	d.Logger = NewLogger(d.LogPrefix, "(~)")
	err = d.Decode("c", &cmd)
	if err == nil {
		d.Logger = NewLogger(d.LogPrefix, fmt.Sprintf("(%d)", cmd))
		return
	}
	if errors.Is(err, io.EOF) {
		err = errEnded
	}
	return
}

func DispatchLocal(d *Dispatcher) (err error) {
	// reject remote requests
	if !d.Authorized {
		return nil
	}
	cmd, err := nextCommand(d)
	if err != nil {
		return
	}
	switch cmd {
	case localAcceptOffer:
		return rpc.Dispatch(d.Conn, "[c]c", d.AcceptOffer)
	case localCreateOffer:
		return rpc.Dispatch(d.Conn, "[c]c [c]c", d.CreateOffer)
	case localListenOffers:
		return rpc.Dispatch(d.Conn, "c", d.ListOffers)
	case localListenStatus:
		return rpc.Dispatch(d.Conn, "c", d.ListenStatus)
	case localListOffers:
		return rpc.Dispatch(d.Conn, "c", d.ListenOffers)
	case localListPeers:
		return rpc.Dispatch(d.Conn, "", d.ListPeers)
	case localUpdatePeer:
		return rpc.Dispatch(d.Conn, "", d.UpdatePeer)
	default:
		return errors.New("protocol violation: unknown command")
	}
}

func DispatchRemote(d *Dispatcher) (err error) {
	cmd, err := nextCommand(d)
	if err != nil {
		return
	}
	switch cmd {
	case cmdClose:
		return errEnded
	case remoteSend:
		return rpc.Dispatch(d.Conn, "", d.Receive)
	case remoteDownload:
		return rpc.Dispatch(d.Conn, "[c]c q q", d.Upload)
	default:
		return errors.New("protocol violation: unknown command")
	}
}

func DispatchInfo(d *Dispatcher) (err error) {
	cmd, err := nextCommand(d)
	if err != nil {
		return
	}
	switch cmd {
	case cmdClose:
		return errEnded
	case infoPing:
		return rpc.Dispatch(d.Conn, "", d.Ping)
	default:
		return errors.New("protocol violation: unknown command")
	}
}
