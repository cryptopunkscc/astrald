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
	logPrefix  string
	callerId   string
	authorized bool

	ctx  context.Context
	api  wrapper.Api
	conn io.ReadWriteCloser

	srv Service
	job *sync.WaitGroup

	cslq *cslq.Endec
	log  *log.Logger
}

func NewDispatcher(
	logPrefix string,
	callerId string,
	authorized bool,
	ctx context.Context,
	api wrapper.Api,
	conn io.ReadWriteCloser,
	srv Service,
	job *sync.WaitGroup,
) *Dispatcher {
	return &Dispatcher{
		logPrefix:  logPrefix,
		callerId:   callerId,
		authorized: authorized,
		ctx:        ctx,
		api:        api,
		conn:       conn,
		srv:        srv,
		job:        job,
		cslq:       cslq.NewEndec(conn),
		log:        NewLogger(logPrefix),
	}
}

func (d Dispatcher) Serve(
	dispatch func(d *Dispatcher) error,
) (err error) {
	for err == nil {
		err = dispatch(&d)
		if err == nil {
			d.log.Println("OK")
		}
	}
	if errors.Is(err, errEnded) {
		d.log.Println("End")
		err = nil
	}
	if err != nil {
		d.log.Println(Error(err, "Failed"))
	}
	return errors.Unwrap(err)
}

func Dispatch(d *Dispatcher) (err error) {
	cmd, err := nextCommand(d)
	if err != nil {
		return
	}
	switch cmd {
	case infoPing:
		return rpc.Dispatch(d.conn, "", d.Ping)
	case remoteSend:
		return rpc.Dispatch(d.conn, "", d.Receive)
	case remoteDownload:
		return rpc.Dispatch(d.conn, "[c]c q q", d.Upload)
	}
	if !d.authorized {
		return nil
	}
	switch cmd {
	case localListPeers:
		return rpc.Dispatch(d.conn, "", d.ListPeers)
	case localCreateOffer:
		return rpc.Dispatch(d.conn, "[c]c [c]c", d.CreateOffer)
	case localAcceptOffer:
		return rpc.Dispatch(d.conn, "[c]c", d.AcceptOffer)
	case localListOffers:
		return rpc.Dispatch(d.conn, "c", d.ListOffers)
	case localListenOffers:
		return rpc.Dispatch(d.conn, "c", d.ListenOffers)
	case localListenStatus:
		return rpc.Dispatch(d.conn, "c", d.ListenStatus)
	case localUpdatePeer:
		return rpc.Dispatch(d.conn, "", d.UpdatePeer)
	}
	return errors.New("protocol violation: unknown command")
}

func nextCommand(d *Dispatcher) (cmd uint8, err error) {
	d.log = NewLogger(d.logPrefix, "(~)")
	err = d.cslq.Decode("c", &cmd)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = errEnded
		}
		return
	}
	d.log = NewLogger(d.logPrefix, fmt.Sprintf("(%d)", cmd))
	if cmd == cmdClose {
		err = errEnded
	}
	return
}
