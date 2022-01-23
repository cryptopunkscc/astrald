package warpdrive

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/mod/id"
	uuid "github.com/nu7hatch/gouuid"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const PORT = "warpdrive"

const (
	SEND   = PORT + "/send"
	ACCEPT = PORT + "/accept"
	REJECT = PORT + "/reject"
)

func (c Config) RunService() {
	service := c.newService()
	service.setupApi()
	service.setupCore()
	service.setupStorage()
	service.setupRepository()
	service.setupResolver()
	service.setupIdentity()
	service.setupPeers()
	service.setupOffers()
	go service.handleServiceSend()
	go service.handleServiceAccept()
	go service.handleServiceReject()
	go service.handleSenderPeers()
	go service.handleSenderSend()
	go service.handleSenderStatus()
	go service.handleSenderSent()
	go service.handleSenderEvents()
	go service.handleRecipientAccept()
	go service.handleRecipientReject()
	go service.handleRecipientOffers()
	go service.handleRecipientReceived()
	go service.handleRecipientUpdate()
	go service.handleRecipientEvents()
	go service.handleCommandLine()
}

func (c Config) newService() *service {
	service := &service{}
	service.Logger = log.Default()
	service.Config = c
	return service
}

type service struct {
	identity string
	core
}

// ================================ Setup =================================

func (srv *service) setupApi() {
	if srv.Api == nil {
		srv.Api = astral.Instance()
	}
}

func (srv *service) setupIdentity() {
	identity, err := id.Query()
	if err != nil {
		srv.Panic("Cannot obtain node identity", err)
	}
	srv.identity = identity.String()
}

func (srv *service) setupPeers() {
	contactList, err := contacts.Query()
	if err != nil {
		srv.Panic("Cannot obtain contacts", err)
	}
	peers := make(Peers)
	for _, contact := range contactList {
		peerId := PeerId(contact.Id)
		peers[peerId] = &Peer{
			Id:    peerId,
			Alias: contact.Name,
		}
	}
	srv.core.setPeers(peers)
	srv.core.setupPeers()
}

// =========================================================================
// ================================ Caller =================================

func (srv service) callServiceSend(peer string, files []Info) (id string, err error) {
	srv.Logger = newLogger("<", SEND)
	// Connect to service
	conn, err := srv.Query(peer, SEND)
	if err != nil {
		srv.Println("Cannot connect", peer, err)
		return
	}
	defer conn.Close()
	// Send file request
	id = srv.newOfferId()
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
	srv.addOutgoingOffer(id, files)
	return
}

func (srv service) newOfferId() string {
	v4, err := uuid.NewV4()
	if err != nil {
		srv.Panic(err)
	}
	return v4.String()
}

func shrinkPaths(in []Info) (out []Info) {
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

// ================================ Handler ================================

func (srv service) handleServiceSend() {
	// Register port
	port := srv.register(SEND)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			peerId := PeerId(request.Caller())
			peer := srv.getPeer(peerId)
			// Check if peer is blocked
			if peer.Mod == PEER_MOD_BLOCK {
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
			var files []Info
			err = dec.Decode(&files)
			if err != nil {
				srv.Println("Cannot read files for offer", offerId, err)
				return
			}
			// Store incoming offer
			srv.addIncomingOffer(peer, offerId, files)
			// Send OK
			_ = enc.Write(conn, uint8(0))
			// Auto accept offer if peer is trusted
			if peer.Mod == PEER_MOD_TRUST {
				_ = srv.callServiceAccept(OfferId(offerId))
			}
		}(request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (srv service) callServiceAccept(id OfferId) (err error) {
	srv.Logger = newLogger("<", ACCEPT)
	// Get cached incoming files by request id
	offer := srv.getIncomingOffer(id)
	if offer == nil {
		err = errors.New(fmt.Sprint("No incoming file with id ", id))
		srv.Println("Cannot find incoming file", err)
		return err
	}
	// Obtain offer reader connection
	filesConn, err := func() (filesConn io.ReadWriteCloser, err error) {
		// Connect to service
		conn, err := srv.Query(string(offer.Peer), ACCEPT)
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
		srv.updateIncomingOfferStatus(offer, "accepted", true)
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
		err = srv.copyFilesFrom(filesConn, offer)
		if err != nil {
			return
		}
		srv.updateIncomingOfferStatus(offer, "downloaded", true)
		// Send OK
		err = enc.Write(filesConn, uint8(0))
		if err != nil {
			srv.Println("Cannot send ok", err)
			return
		}
	}()
	return
}

// ================================ Handler ================================

func (srv service) handleServiceAccept() {
	// Register port
	port := srv.register(ACCEPT)
	for request := range port.Next() {
		go func(request astral.Request) {
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
			offer := srv.getOutgoingOffer(OfferId(offerId))
			if offer == nil {
				srv.Println("Cannot find offer with id", offerId, err)
				conn.Close()
			}
			// Update status
			srv.updateOutgoingOfferStatus(offer, "accepted", true)
			// Register port for reading files
			filesQuery := PORT + "/" + string(offer.Id)
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
			err = srv.copyFilesTo(filesConn, offer)
			if err != nil {
				return
			}
			// Wait OK
			_, err = enc.ReadUint8(filesConn)
			if err != nil {
				srv.Println("Cannot read ok", filesQuery, err)
				return
			}
			srv.updateOutgoingOfferStatus(offer, "uploaded", true)
		}(request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (srv service) callServiceReject(id OfferId) (err error) {
	srv.Logger = newLogger("<", REJECT)
	// Get cached incoming files by id
	offer := srv.getIncomingOffer(id)
	if offer == nil {
		err = errors.New(fmt.Sprint("No incoming file with id ", id))
		srv.Println("Cannot find incoming file", err)
		return
	}
	// Connect to service
	conn, err := srv.Query(string(offer.Peer), REJECT)
	if err != nil {
		srv.Println("Cannot connect", err)
		return
	}
	defer conn.Close()
	// Send rejected offer id
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		srv.Println("Cannot send rejected offer id", id, err)
		return
	}
	// Wait for OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		srv.Println("Cannot read ok", id, err)
		return
	}
	srv.updateIncomingOfferStatus(offer, "rejected", true)
	return
}

// ================================ Handler ================================

func (srv service) handleServiceReject() {
	// Register port
	port := srv.register(REJECT)
	for request := range port.Next() {
		go func(request astral.Request) {
			// Accept incoming connection
			conn, err := request.Accept()
			if err != nil {
				srv.Println("Cannot accept connection from", request.Caller(), err)
				return
			}
			defer conn.Close()
			// Read id of rejected outgoing files
			offerId, err := enc.ReadL8String(conn)
			if err != nil {
				srv.Println("Cannot read request id", request.Caller(), err)
				return
			}
			// Reject outgoing files
			offer := srv.getOutgoingOffer(OfferId(offerId))
			if offer != nil {
				srv.updateOutgoingOfferStatus(offer, "rejected", true)
			}
			// Send OK
			err = enc.Write(conn, uint8(0))
			if err != nil {
				srv.Println("Cannot send ok", request.Caller(), err)
				return
			}
		}(request)
	}
}

// =========================================================================
// ================================ Utils ================================

func (srv *service) register(query string) (port astral.Port) {
	srv.Logger = newLogger(">", query)
	port, err := srv.Register(query)
	if err != nil {
		srv.Panicln("Cannot register port", query, err)
	}
	go func() {
		<-srv.Done()
		_ = port.Close()
	}()
	return
}

func newLogger(prefix ...string) *log.Logger {
	var chunks []interface{}
	suffix := " "
	for i, chunk := range prefix {
		if i == len(prefix)-1 {
			suffix = ": "
		}
		chunks = append(chunks, chunk+suffix)
	}
	return log.New(os.Stderr, fmt.Sprint(chunks...), log.LstdFlags|log.Lmsgprefix)
}
