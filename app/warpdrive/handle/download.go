package handle

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/legacy/enc"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
	"time"
)

func (c Client) Download(id api.OfferId) (err error) {
	// Connect to local service
	conn, err := c.query(api.QueryAccept)
	if err != nil {
		return
	}
	defer conn.Close()
	// Send accepted request id to service
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		c.Println("Cannot send accepted request id", err)
		return
	}
	// Read OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		c.Println("Cannot read ok", err)
		return
	}
	return
}

func Download(ctx handler.Context, request astral.Request) {
	if ctx.IsRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		ctx.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	id, err := enc.ReadL8String(conn)
	if err != nil {
		ctx.Println("Cannot read request id", err)
		return
	}
	err = download(ctx, api.OfferId(id))
	if err != nil {
		ctx.Println("Cannot download incoming files", id, err)
		return
	}
	err = enc.Write(conn, uint8(0))
	if err != nil {
		ctx.Println("Cannot send ok", err)
		return
	}
	ctx.Println("Accepted incoming files", id)
}

func download(ctx handler.Context, id api.OfferId) (err error) {
	ctx.LogPrefix("<", api.QueryFiles)
	// Get incoming offer service for offer id
	offer := service.Incoming(ctx.Core).Set(id)
	if offer.Offer == nil {
		err = errors.New(fmt.Sprint("No incoming file with id ", id))
		ctx.Println("Cannot find incoming file", err)
		return err
	}
	// Update status
	offer.Accept()
	// Obtain offer reader connection
	filesConn, err := func() (filesConn io.ReadWriteCloser, err error) {
		peerId := string(offer.Peer)
		// Connect to service
		conn, err := ctx.Query(peerId, api.QueryFiles)
		if err != nil {
			ctx.Println("Cannot connect", err)
			return
		}
		defer conn.Close()
		// Send file request id
		err = enc.WriteL8String(conn, string(offer.Id))
		if err != nil {
			ctx.Println("Cannot send request id", err)
			return
		}
		// Read name of port for downloading files
		filesQuery, err := enc.ReadL8String(conn)
		if err != nil {
			ctx.Println("Cannot read files port", err)
			return
		}
		// Send ok
		err = enc.Write(conn, uint8(0))
		if err != nil {
			ctx.Println("Cannot send ok", err)
			return
		}
		// Open connection for downloading files
		filesConn, err = ctx.Query(peerId, filesQuery)
		if err != nil {
			ctx.Println("Cannot query files port", err)
			return
		}
		return
	}()
	if err != nil {
		return err
	}
	// Try to download files in background
	go func() {
		defer func() {
			_ = filesConn.Close()
			// Ensure the status will be updated
			time.Sleep(200 * time.Millisecond)
			offer.Finish(err)
		}()
		// Copy files to storage
		err = service.File(ctx.Core).Copy(offer).From(filesConn)
		if err != nil {
			return
		}
		// Send OK
		err = enc.Write(filesConn, uint8(0))
		if err != nil {
			ctx.Println("Cannot send ok", err)
			return
		}
	}()
	return
}

func Upload(ctx handler.Context, request astral.Request) {
	// Accept incoming connection
	conn, err := request.Accept()
	if err != nil {
		ctx.Println("Cannot accept connection from", request.Caller(), err)
		return
	}
	defer conn.Close()
	// Read request id
	id, err := enc.ReadL8String(conn)
	if err != nil {
		ctx.Println("Cannot read offer id", err)
		return
	}

	offer := service.Outgoing(ctx.Core).Set(api.OfferId(id))
	// Obtain file by request id
	if offer.Offer == nil {
		ctx.Println("Cannot find offer with id", id)
		return
	}
	// Update status
	offer.Accept()
	// Register port for reading files
	filesQuery := api.Port + "/" + string(offer.Id)
	filesPort, err := ctx.Register(filesQuery)
	if err != nil {
		ctx.Println("Cannot register files port", filesPort, err)
		return
	}
	defer filesPort.Close()
	// Send query port to recipient
	err = enc.WriteL8String(conn, filesQuery)
	if err != nil {
		ctx.Println("Cannot send files port", filesQuery, err)
		return
	}
	// Read OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		ctx.Println("Cannot read ok", filesQuery, err)
		return
	}
	// Wait for connection on files port
	filesRequest := <-filesPort.Next()
	if filesRequest.Caller() != request.Caller() {
		filesRequest.Reject()
		ctx.Println("Invalid caller", filesQuery, err)
		return
	}
	filesConn, err := filesRequest.Accept()
	if err != nil {
		ctx.Println("Cannot accept files connection", filesQuery, err)
		return
	}
	defer func() {
		_ = filesConn.Close()
		time.Sleep(200 * time.Millisecond)
		offer.Finish(err)
	}()
	go func() {
		// Send files
		err = service.File(ctx.Core).Copy(offer).To(filesConn)
		if err != nil {
			ctx.Println("Cannot copy files", filesQuery, err)
			return
		}
	}()
	// Wait OK
	_, err = enc.ReadUint8(filesConn)
	if err != nil {
		ctx.Println("Cannot read ok", filesQuery, err)
		return
	}
}
