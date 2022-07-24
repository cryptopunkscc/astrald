package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
	uuid "github.com/nu7hatch/gouuid"
	"log"
	"net/url"
	"path/filepath"
	"strings"
)

func (c Client) Send(peerId api.PeerId, filePath string) (id api.OfferId, accepted bool, err error) {
	// Connect to local service
	conn, err := c.query(api.QuerySend)
	if err != nil {
		return
	}
	defer conn.Close()
	// Send recipient id
	err = cslq.Encode(conn, "[c]c", peerId)
	if err != nil {
		c.Println("Cannot send recipient id", err)
		return
	}
	// Send file path
	err = cslq.Encode(conn, "[c]c", filePath)
	if err != nil {
		c.Println("Cannot send file path", err)
		return
	}
	// Read offer id
	err = cslq.Decode(conn, "[c]c", &id)
	if err != nil {
		c.Println("Cannot read offer id", err)
		return
	}
	// Read result code
	var code byte
	err = cslq.Decode(conn, "c", &code)
	if err != nil {
		c.Println("Cannot read offer result code", err)
	}
	accepted = code == 1
	return
}

func Send(srv handler.Context, request astral.Request) {
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
	// Read peer id
	var peerId api.PeerId
	err = cslq.Decode(conn, "[c]c", &peerId)
	if err != nil {
		srv.Println("Cannot read peer id", err)
		return
	}
	// Read file path
	var filePath string
	err = cslq.Decode(conn, "[c]c", &filePath)
	if err != nil {
		srv.Println("Cannot read file path", err)
		return
	}
	// Get files info
	files, err := service.File(srv.Core).Info(filePath)
	if err != nil {
		srv.Println("Cannot get files info", err)
		return
	}
	// Send file to recipient service
	id, code, err := send(srv, peerId, files)
	if err != nil {
		srv.Println("Cannot send file", err)
		return
	}
	// Write id to sender
	err = cslq.Encode(conn, "[c]c", id)
	if err != nil {
		srv.Println("Cannot send id", id, err)
		return
	}
	// Write code to sender
	err = cslq.Encode(conn, "c", code)
	if err != nil {
		srv.Println("Cannot code", id, err)
		return
	}
	srv.Println(filePath, "offer sent to", peerId)
}

func send(srv handler.Context, peer api.PeerId, files []api.Info) (id api.OfferId, code uint8, err error) {
	srv.LogPrefix("<", api.QueryOffer)
	// Connect to service
	conn, err := srv.Query(string(peer), api.QueryOffer)
	if err != nil {
		srv.Println("Cannot connect", peer, len(peer), err)
		return
	}
	defer conn.Close()
	// Send file request
	id = newOfferId()
	err = cslq.Encode(conn, "[c]c", id)
	if err != nil {
		srv.Println("Cannot send offer id", peer, err)
		return
	}
	shrunken := shrinkPaths(files)
	err = json.NewEncoder(conn).Encode(shrunken)
	if err != nil {
		srv.Println("Cannot send offer info", id, peer, err)
		return
	}
	service.Outgoing(srv.Core).Add(id, files, peer)
	// Read result code
	err = cslq.Decode(conn, "c", &code)
	if err != nil {
		srv.Println("Cannot read result code", peer, err)
		return
	}
	return
}

func newOfferId() api.OfferId {
	v4, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return api.OfferId(v4.String())
}

// TODO make it bulletproof
func shrinkPaths(in []api.Info) (out []api.Info) {
	dir, _ := filepath.Split(in[0].Uri)
	if dir == "" {
		return in
	}
	uri, err := url.Parse(dir)
	if err != nil {
		log.Println("Cannot parse uri", err)
		return in
	}
	if uri.Scheme != "" {
		for _, info := range in {
			uri, err = url.Parse(info.Uri)
			if err != nil {
				log.Println("Cannot parse uri", err)
				return in
			}
			path := strings.Replace(uri.Path, ":", "/", -1)
			_, file := filepath.Split(path)
			info.Uri = file
			out = append(out, info)
		}
	} else {
		for _, info := range in {
			info.Uri = strings.TrimPrefix(info.Uri, dir)
			out = append(out, info)
		}
	}
	return
}

func Receive(srv handler.Context, request astral.Request) {
	peerId := api.PeerId(request.Caller())
	peer := service.Peer(srv.Core).Get(peerId)
	// Check if peer is blocked
	if peer.Mod == api.PeerModBlock {
		request.Reject()
		srv.Println("Blocked request from", peerId)
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request from", peerId, err)
		return
	}
	defer conn.Close()
	var offerId api.OfferId
	err = cslq.Decode(conn, "[c]c", &offerId)
	if err != nil {
		srv.Println("Cannot read offer id", err)
	}
	// Read files request
	dec := json.NewDecoder(conn)
	var files []api.Info
	err = dec.Decode(&files)
	if err != nil {
		srv.Println("Cannot read files for offer", offerId, err)
		return
	}
	// Store incoming offer
	service.Incoming(srv.Core).Add(offerId, files, peerId)
	// Auto accept offer if peer is trusted
	code := api.OfferAwaiting
	if peer.Mod == api.PeerModTrust {
		err = download(srv, api.OfferId(offerId))
		if err != nil {
			srv.Println("Cannot auto accept files offer", offerId, err)
		} else {
			code = api.OfferAccepted
		}
	}
	// Send received
	_ = cslq.Encode(conn, "c", code)
}
