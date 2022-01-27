package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	uuid "github.com/nu7hatch/gouuid"
	"path/filepath"
	"strings"
)

func (s Sender) Send(peerId api.PeerId, filePath string) (id api.OfferId, err error) {
	// Connect to local service
	conn, err := s.query(api.SenSend)
	if err != nil {
		return
	}
	defer conn.Close()
	// Send recipient id
	err = enc.WriteL8String(conn, string(peerId))
	if err != nil {
		s.Println("Cannot send recipient id", err)
		return
	}
	// Send file path
	err = enc.WriteL8String(conn, filePath)
	if err != nil {
		s.Println("Cannot send file path", err)
		return
	}
	// Read response
	strId, err := enc.ReadL8String(conn)
	if err != nil {
		s.Println("Cannot read response id", err)
	}
	id = api.OfferId(strId)
	return
}

func SenderSend(srv handler.Context, request astral.Request) {
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
	peerId, err := enc.ReadL8String(conn)
	if err != nil {
		srv.Println("Cannot read peer id", err)
		return
	}
	// Read file path
	filePath, err := enc.ReadL8String(conn)
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
	id, err := send(srv, peerId, files)
	if err != nil {
		srv.Println("Cannot send file", err)
		return
	}
	// Write response to sender
	err = enc.WriteL8String(conn, id)
	if err != nil {
		srv.Println("Cannot send id", id, err)
		return
	}
	srv.Println(filePath, "offer sent to", peerId)
}

func send(srv handler.Context, peer string, files []api.Info) (id string, err error) {
	srv.LogPrefix("<", api.Send)
	// Connect to service
	conn, err := srv.Query(peer, api.Send)
	if err != nil {
		srv.Println("Cannot connect", peer, err)
		return
	}
	defer conn.Close()
	// Send file request
	id = newOfferId()
	err = enc.WriteL8String(conn, id)
	if err != nil {
		srv.Println("Cannot send offer id", peer, err)
		return "", err
	}
	shrunken := shrinkPaths(files)
	err = json.NewEncoder(conn).Encode(shrunken)
	if err != nil {
		srv.Println("Cannot send offer info", id, peer, err)
		return
	}
	// Wait for close
	_, err = enc.ReadUint8(conn)
	if err != nil {
		srv.Println("Cannot read ok", peer, err)
		return
	}
	service.Outgoing(srv.Core).Add(id, files, nil)
	return
}

func newOfferId() string {
	v4, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return v4.String()
}

func shrinkPaths(in []api.Info) (out []api.Info) {
	dir, _ := filepath.Split(in[0].Path)
	if dir == "" {
		return in
	}
	for _, info := range in {
		info.Path = strings.TrimPrefix(info.Path, dir)
		out = append(out, info)
	}
	return
}

func ServiceSend(srv handler.Context, request astral.Request) {
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
	offerId, err := enc.ReadL8String(conn)
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
	service.Incoming(srv.Core).Add(offerId, files, &peer)
	// Send OK
	_ = enc.Write(conn, uint8(0))
	// Auto accept offer if peer is trusted
	if peer.Mod == api.PeerModTrust {
		_ = accept(srv, api.OfferId(offerId))
	}
}
