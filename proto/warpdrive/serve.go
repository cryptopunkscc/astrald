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

func Dispatch(d *Dispatcher) (err error) {
	d.Logger = NewLogger(d.LogPrefix, "(~)")
	var cmd uint8
	err = d.Decode("c", &cmd)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = errEnded
		}
		return
	}
	d.Logger = NewLogger(d.LogPrefix, fmt.Sprintf("(%d)", cmd))
	switch cmd {
	case cmdClose:
		return errEnded
	case remoteSend:
		return rpc.Dispatch(d.Conn, "", d.Receive)
	case remoteDownload:
		return rpc.Dispatch(d.Conn, "[c]c", d.Upload)
	default:
		// reject remote requests
		if !d.Authorized {
			return nil
		}
		switch cmd {
		case localAccept:
			return rpc.Dispatch(d.Conn, "[c]c", d.Accept)
		case localSend:
			return rpc.Dispatch(d.Conn, "[c]c [c]c", d.send)
		case localOffers:
			return rpc.Dispatch(d.Conn, "[c]c", d.offers)
		case localStatus:
			return rpc.Dispatch(d.Conn, "[c]c", d.status)
		case localSubscribe:
			return rpc.Dispatch(d.Conn, "[c]c", d.subscribe)
		case localPeers:
			return rpc.Dispatch(d.Conn, "", d.peers)
		case localUpdate:
			return rpc.Dispatch(d.Conn, "", d.update)
		case localPing:
			return rpc.Dispatch(d.Conn, "", d.Ping)
		default:
			return errors.New("protocol violation: unknown command")
		}
	}
}
