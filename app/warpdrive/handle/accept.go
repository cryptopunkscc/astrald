package handle

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
)

func (r Recipient) Accept(id api.OfferId) (err error) {
	// Connect to local service
	conn, err := r.query(api.RecAccept)
	if err != nil {
		return
	}
	defer conn.Close()
	// Send accepted request id to service
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		r.Println("Cannot send accepted request id", err)
		return
	}
	// Read OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		r.Println("Cannot read ok", err)
		return
	}
	return
}

func RecipientAccept(srv handler.Context, request astral.Request) {
	if srv.IsRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	id, err := enc.ReadL8String(conn)
	if err != nil {
		srv.Println("Cannot read request id", err)
		return
	}
	err = accept(srv, api.OfferId(id))
	if err != nil {
		srv.Println("Cannot accept incoming files", id, err)
		return
	}
	err = enc.Write(conn, uint8(0))
	if err != nil {
		srv.Println("Cannot send ok", err)
		return
	}
	srv.Println("Accepted incoming files", id)
}

func accept(srv handler.Context, id api.OfferId) (err error) {
	srv.LogPrefix("<", api.Accept)
	// Get cached incoming files by request id
	offer := service.Incoming(srv.Core).Get()[id]
	if offer == nil {
		err = errors.New(fmt.Sprint("No incoming file with id ", id))
		srv.Println("Cannot find incoming file", err)
		return err
	}
	// Obtain offer reader connection
	filesConn, err := func() (filesConn io.ReadWriteCloser, err error) {
		// Connect to service
		conn, err := srv.Query(string(offer.Peer), api.Accept)
		if err != nil {
			srv.Println("Cannot connect", err)
			return
		}
		defer conn.Close()
		// Send file request id
		err = enc.WriteL8String(conn, string(offer.Id))
		if err != nil {
			srv.Println("Cannot send request id", err)
			return
		}
		// Read name of port for downloading files
		filesQuery, err := enc.ReadL8String(conn)
		if err != nil {
			srv.Println("Cannot read files port", err)
			return
		}
		// Update status
		offer.Status.Status = api.StatusAccepted
		service.Incoming(srv.Core).Update(offer, -1)
		// Send ok
		err = enc.Write(conn, uint8(0))
		if err != nil {
			srv.Println("Cannot send ok", err)
			return
		}
		// Open connection for downloading files
		filesConn, err = srv.Query(string(offer.Peer), filesQuery)
		if err != nil {
			srv.Println("Cannot query files port", err)
			return
		}
		return
	}()
	if err != nil {
		return err
	}
	// Try to download files in background
	go func() {
		defer filesConn.Close()
		// Copy files to storage
		err = service.File(srv.Core).CopyFrom(filesConn, offer)
		if err != nil {
			return
		}
		offer.Status.Status = api.StatusCompleted
		service.Incoming(srv.Core).Update(offer, -1)
		// Send OK
		err = enc.Write(filesConn, uint8(0))
		if err != nil {
			srv.Println("Cannot send ok", err)
			return
		}
	}()
	return
}

func ServiceAccept(srv handler.Context, request astral.Request) {
	// Accept incoming connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept connection from", request.Caller(), err)
		return
	}
	defer conn.Close()
	// Read request id
	offerId, err := enc.ReadL8String(conn)
	if err != nil {
		srv.Println("Cannot read offer id", err)
		return
	}
	// Obtain file by request id
	offer := service.Outgoing(srv.Core).Get()[api.OfferId(offerId)]
	if offer == nil {
		srv.Println("Cannot find offer with id", offerId, err)
		return
	}
	// Update status
	offer.Status.Status = api.StatusAccepted
	service.Outgoing(srv.Core).Update(offer, -1)
	// Register port for reading files
	filesQuery := api.Port + "/" + string(offer.Id)
	filesPort, err := srv.Register(filesQuery)
	if err != nil {
		srv.Println("Cannot register files port", filesPort, err)
		return
	}
	defer filesPort.Close()
	// Send query port to recipient
	err = enc.WriteL8String(conn, filesQuery)
	if err != nil {
		srv.Println("Cannot send files port", filesQuery, err)
		return
	}
	// Read OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		srv.Println("Cannot read ok", filesQuery, err)
		return
	}
	// Wait for connection on files port
	filesRequest := <-filesPort.Next()
	if filesRequest.Caller() != request.Caller() {
		filesRequest.Reject()
		srv.Println("Invalid caller", filesQuery, err)
		return
	}
	filesConn, err := filesRequest.Accept()
	if err != nil {
		srv.Println("Cannot accept files connection", filesQuery, err)
		return
	}
	defer filesConn.Close()
	// Send files
	err = service.File(srv.Core).CopyTo(filesConn, offer)
	if err != nil {
		return
	}
	// Wait OK
	_, err = enc.ReadUint8(filesConn)
	if err != nil {
		srv.Println("Cannot read ok", filesQuery, err)
		return
	}
	offer.Status.Status = api.StatusCompleted
	service.Outgoing(srv.Core).Update(offer, -1)
}
