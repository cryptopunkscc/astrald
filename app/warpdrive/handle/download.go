package handle

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/cslq"
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
	err = cslq.Encode(conn, "[c]c", id)
	if err != nil {
		c.Println("Cannot send accepted request id", err)
		return
	}
	// Read OK
	var code byte
	err = cslq.Decode(conn, "c", &code)
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
	// Download offer
	var id api.OfferId
	err = cslq.Decode(conn, "[c]c", &id)
	if err != nil {
		ctx.Println("Cannot read request id", err)
		return
	}
	ctx.Println("Accepted incoming files", id)
	err = download(ctx, id)
	if err != nil {
		ctx.Println("Cannot download incoming files", id, err)
		return
	}
	// Send ok
	err = cslq.Encode(conn, "c", 0)
	if err != nil {
		ctx.Println("Cannot send ok", err)
		return
	}
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
	peerId := string(offer.Peer)

	// Connect to service
	conn, err := ctx.Query(peerId, api.QueryFiles)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		ctx.Println("Cannot connect", err)
		return
	}

	// Send file request id
	err = cslq.Encode(conn, "[c]c", offer.Id)
	if err != nil {
		ctx.Println("Cannot send request id", err)
		return
	}

	// Read confirmation
	var code byte
	err = cslq.Decode(conn, "c", &code)
	if err != nil {
		ctx.Println("Cannot read confirmation", err)
		return
	}

	// download files in background
	go func(conn io.ReadWriteCloser) {
		// Ensure the status will be updated
		var err error
		defer func() {
			conn.Close()
			time.Sleep(200 * time.Millisecond)
			offer.Finish(err)
		}()

		// Copy files to storage
		err = service.File(ctx.Core).Copy(offer).From(conn)
		if err != nil {
			ctx.Println("Cannot download files", err)
			return
		}

		// Send OK
		err = cslq.Encode(conn, "c", 0)
		if err != nil {
			ctx.Println("Cannot send ok", err)
			return
		}
	}(conn)
	conn = nil
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
	var id api.OfferId
	err = cslq.Decode(conn, "[c]c", &id)
	if err != nil {
		ctx.Println("Cannot read offer id", err)
		return
	}
	offer := service.Outgoing(ctx.Core).Set(id)

	// Obtain file by request id
	if offer.Offer == nil {
		ctx.Println("Cannot find offer with id", id)
		return
	}

	// Update status
	offer.Accept()

	// Send confirmation
	err = cslq.Encode(conn, "c", 0)
	if err != nil {
		ctx.Println("Cannot send confirmation", err)
		return
	}

	// Ensure the status will be updated
	defer func() {
		time.Sleep(200 * time.Millisecond)
		offer.Finish(err)
	}()

	// Send files
	err = service.File(ctx.Core).Copy(offer).To(conn)
	if err != nil {
		ctx.Println("Cannot upload files", err)
		return
	}

	// Read OK
	var code byte
	err = cslq.Decode(conn, "c", &code)
	if err != nil {
		ctx.Println("Cannot read ok", err)
		return
	}
}
