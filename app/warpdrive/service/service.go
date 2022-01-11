package warpdrive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/mod/id"
	"github.com/mitchellh/ioprogress"
	uuid "github.com/nu7hatch/gouuid"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const PORT = "warpdrive"

const (
	SEND   = PORT + "/send"
	ACCEPT = PORT + "/accept"
	REJECT = PORT + "/reject"
)

func RunService(ctx context.Context) {
	service := newService(ctx)
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

func newService(ctx context.Context) *service {
	service := &service{
		ctx:      ctx,
		peers:    Peers{},
		home:     userFiles(),
		received: receivedFiles(),
		repo:     newRepository(),
		offers: offers{
			incoming: Offers{},
			outgoing: Offers{},
		},
	}
	service.repo.init()
	service.setupIdentity()
	service.setupPeers()
	service.setupOffers()
	return service
}

type service struct {
	ctx      context.Context
	identity string
	home     storage
	received storage
	peers    Peers
	repo     repository
	offers
	notify
}

type offers struct {
	incoming Offers
	outgoing Offers
}

type notify struct {
	mu             sync.Mutex
	filesRequest   []io.WriteCloser
	incomingStatus []io.WriteCloser
	outgoingStatus []io.WriteCloser
}

func (srv *service) setupIdentity() {
	identity, err := id.Query()
	if err != nil {
		log.Panic("Cannot obtain node identity", err)
	}
	srv.identity = identity.String()
}

func (srv *service) setupPeers() {
	contactList, err := contacts.Query()
	if err != nil {
		log.Panic("Cannot obtain contacts", err)
	}
	for _, contact := range contactList {
		peerId := PeerId(contact.Id)
		srv.peers[peerId] = &Peer{
			Id:    peerId,
			Alias: contact.Name,
		}
	}
	peers := srv.repo.listPeers()
	for _, peer := range peers {
		peerRef := peer
		srv.peers[peer.Id] = &peerRef
	}
}

func (srv *service) setupOffers() {
	srv.incoming = srv.repo.listIncoming()
	srv.outgoing = srv.repo.listOutgoing()
}

// =========================================================================
// ================================ Caller =================================

func (srv *service) callServiceSend(peer string, files []Info) (id string, err error) {
	// Connect to service
	conn, err := astral.Query(peer, SEND)
	if err != nil {
		log.Println("<", SEND, "Cannot connect", peer, err)
		return
	}
	defer conn.Close()
	// Send file request
	id = newOfferId()
	err = enc.WriteL8String(conn, id)
	if err != nil {
		log.Println("<", SEND, "Cannot send offer id", peer, err)
		return "", err
	}
	shrunken := shrinkPaths(files)
	err = json.NewEncoder(conn).Encode(shrunken)
	if err != nil {
		log.Println("<", SEND, "Cannot send offer info", id, peer, err)
		return
	}
	// Wait for close
	_, err = enc.ReadUint8(conn)
	if err != nil {
		log.Println("<", SEND, "Cannot read ok", peer, err)
		return
	}
	// Cache outgoing files request
	offer := &Offer{
		Status: Status{
			Id:     OfferId(id),
			Status: "sent",
		},
		Files: files,
	}
	srv.outgoing[offer.Id] = offer
	srv.repo.saveOutgoing(offer)
	// Notify status listeners
	go notifyListeners("<", SEND, offer.Status, srv.outgoingStatus)
	return
}

func newOfferId() string {
	v4, err := uuid.NewV4()
	if err != nil {
		log.Panic(err)
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

func (srv *service) handleServiceSend() {
	// Register port
	port := srv.register(SEND)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			peerId := PeerId(request.Caller())
			peerMode := srv.peerMod(peerId)
			// Check if peer is blocked
			if peerMode == PEER_MOD_BLOCK {
				request.Reject()
				log.Println(">", SEND, "Blocked request from", peerId)
				return
			}
			// Accept connection
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", SEND, "Cannot accept request from", peerId, err)
				return
			}
			defer conn.Close()
			offerId, err := enc.ReadL8String(conn)
			if err != nil {
				log.Println(">", SEND, "Cannot read offer id", err)
			}
			// Read files request
			dec := json.NewDecoder(conn)
			var files []Info
			err = dec.Decode(&files)
			if err != nil {
				log.Println(">", SEND, "Cannot read files for offer", offerId, err)
				return
			}
			// Save incoming offer
			offer := &Offer{
				Files: files,
				Peer:  peerId,
				Status: Status{
					Id:     OfferId(offerId),
					Status: "received",
				},
			}
			srv.incoming[offer.Id] = offer
			srv.repo.saveIncoming(offer)
			// Send OK
			_ = enc.Write(conn, uint8(0))
			// Notify status listeners
			go notifyListeners(">", SEND, offer.Status, srv.incomingStatus)
			// Auto accept incoming offer if peer is trusted
			switch peerMode {
			case PEER_MOD_TRUST:
				_ = srv.callServiceAccept(offer.Id)
			case PEER_MOD_ASK:
				go notifyListeners(">", SEND, offer, srv.filesRequest)
			}
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (srv *service) callServiceAccept(id OfferId) (err error) {
	// Get cached incoming files by request id
	offer := srv.incoming[id]
	if offer == nil {
		err = errors.New(fmt.Sprint("No incoming file with id ", id))
		log.Println("<", ACCEPT, "Cannot find incoming file", err)
		return err
	}
	// Obtain offer reader connection
	filesConn, err := func() (filesConn io.ReadWriteCloser, err error) {
		// Connect to service
		conn, err := astral.Query(string(offer.Peer), ACCEPT)
		if err != nil {
			log.Println("<", ACCEPT, "Cannot connect", err)
			return
		}
		defer conn.Close()
		// Send file request id
		err = enc.WriteL8String(conn, string(offer.Id))
		if err != nil {
			log.Println("<", ACCEPT, "Cannot send request id", err)
			return
		}
		// Read name of port for downloading files
		filesQuery, err := enc.ReadL8String(conn)
		if err != nil {
			log.Println("<", ACCEPT, "Cannot read files port", err)
			return
		}
		offer.Status.Status = "accepted"
		srv.repo.saveIncoming(offer)
		go notifyListeners("<", ACCEPT, offer.Status, srv.incomingStatus)
		err = enc.Write(conn, uint8(0))
		if err != nil {
			log.Println("<", ACCEPT, "Cannot send ok", err)
			return
		}
		// Open connection for downloading files
		filesConn, err = astral.Query(string(offer.Peer), filesQuery)
		if err != nil {
			log.Println("<", ACCEPT, "Cannot query files port", err)
			return
		}
		return
	}()
	if err != nil {
		return err
	}
	// Try to download files in background
	go func() {
		// Copy files to storage
		defer filesConn.Close()
		for _, file := range offer.Files {
			err = func() error {
				// Obtain writer
				if file.IsDir {
					err := srv.received.MkDir(file.Path, file.Perm)
					if err != nil && !os.IsExist(err) {
						log.Println("<", ACCEPT, "Cannot make dir", file.Path, err)
						return err
					}
				} else {
					writer, err := srv.received.Writer(file.Path, file.Perm)
					if err != nil {
						log.Println("<", ACCEPT, "Cannot get writer for", file.Path, err)
						return err
					}
					defer writer.Close()
					// Copy bytes
					progress := &ioprogress.Reader{
						Reader:       filesConn,
						Size:         file.Size,
						DrawInterval: 200 * time.Millisecond,
						DrawFunc: func(progress int64, size int64) error {
							offer.Status.Status = fmt.Sprintf("download: %s %d/%dB", file.Path, progress, size)
							go notifyListeners("<", ACCEPT, offer.Status, srv.incomingStatus)
							return nil
						},
					}
					_, err = io.CopyN(writer, progress, file.Size)
					if err != nil {
						log.Println("<", ACCEPT, "Cannot copy", file.Path, err)
						return err
					}
					err = writer.Close()
					if err != nil {
						log.Println("<", ACCEPT, "Cannot close file", file.Path, err)
						return err
					}
				}
				return err
			}()
			if err != nil {
				return
			}
		}
		offer.Status.Status = "downloaded"
		srv.repo.saveIncoming(offer)
		go notifyListeners("<", ACCEPT, offer.Status, srv.incomingStatus)
		// Send OK
		err = enc.Write(filesConn, uint8(0))
		if err != nil {
			log.Println("<", ACCEPT, "Cannot send ok", err)
			return
		}
	}()
	return
}

// ================================ Handler ================================

func (srv *service) handleServiceAccept() {
	// Register port
	port := srv.register(ACCEPT)
	for request := range port.Next() {
		go func(request astral.Request) {
			// Accept incoming connection
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", ACCEPT, "Cannot accept connection from", request.Caller(), err)
				return
			}
			defer conn.Close()
			// Read request id
			offerId, err := enc.ReadL8String(conn)
			if err != nil {
				log.Println(">", ACCEPT, "Cannot read offer id", err)
				return
			}
			// Obtain file by request id
			offer := srv.outgoing[OfferId(offerId)]
			if offer == nil {
				log.Println(">", ACCEPT, "Cannot read offer id", err)
				conn.Close()
			}
			offer.Status.Status = "accepted"
			srv.repo.saveOutgoing(offer)
			go notifyListeners(">", ACCEPT, offer.Status, srv.outgoingStatus)
			// Register port for reading files
			filesQuery := PORT + "/" + string(offer.Id)
			filesPort, err := astral.Reqister(filesQuery)
			if err != nil {
				log.Println(">", ACCEPT, "Cannot register files port", filesPort, err)
				return
			}
			defer filesPort.Close()
			// Send query port to recipient
			err = enc.WriteL8String(conn, filesQuery)
			if err != nil {
				log.Println(">", ACCEPT, "Cannot send files port", filesQuery, err)
				return
			}
			// Read OK
			_, err = enc.ReadUint8(conn)
			if err != nil {
				log.Println(">", ACCEPT, "Cannot read ok", filesQuery, err)
				return
			}
			// Wait for connection on files port
			filesRequest := <-filesPort.Next()
			if filesRequest.Caller() != request.Caller() {
				filesRequest.Reject()
				log.Println(">", ACCEPT, "Invalid caller", filesQuery, err)
				return
			}
			filesConn, err := filesRequest.Accept()
			if err != nil {
				log.Println(">", ACCEPT, "Cannot accept files connection", filesQuery, err)
				return
			}
			defer filesConn.Close()
			// Send files
			for _, file := range offer.Files {
				if !file.IsDir {
					reader, err := srv.home.Reader(file.Path)
					if err != nil {
						log.Println(">", ACCEPT, "Cannot get reader", file.Path, offerId, err)
						return
					}
					progress := &ioprogress.Reader{
						Reader:       reader,
						Size:         file.Size,
						DrawInterval: 200 * time.Millisecond,
						DrawFunc: func(progress int64, size int64) error {
							offer.Status.Status = fmt.Sprintf("upload %s %d/%dB", file.Path, progress, size)
							go notifyListeners(">", ACCEPT, offer.Status, srv.outgoingStatus)
							return nil
						},
					}
					_, err = io.CopyN(filesConn, progress, file.Size)
					if err != nil {
						log.Println(">", ACCEPT, "Cannot send", file.Path, "to", filesRequest.Caller(), err)
						return
					}
				}
			}
			// Wait OK
			_, err = enc.ReadUint8(filesConn)
			if err != nil {
				log.Println(">", ACCEPT, "Cannot read ok", filesQuery, err)
				return
			}
			offer.Status.Status = "uploaded"
			srv.repo.saveOutgoing(offer)
			go notifyListeners(">", ACCEPT, offer.Status, srv.outgoingStatus)
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (srv *service) callServiceReject(id OfferId) (err error) {
	// Get cached incoming files by id
	offer := srv.incoming[id]
	if offer == nil {
		err = errors.New(fmt.Sprint("No incoming file with id ", id))
		log.Println("<", REJECT, "Cannot find incoming file", err)
		return
	}
	// Connect to service
	conn, err := astral.Query(string(offer.Peer), REJECT)
	if err != nil {
		log.Println("<", REJECT, "Cannot connect", err)
		return
	}
	defer conn.Close()
	// Send rejected offer id
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		log.Println("<", REJECT, "Cannot send rejected offer id", id, err)
		return
	}
	// Wait for OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		log.Println("<", REJECT, "Cannot read ok", id, err)
		return
	}
	offer.Status.Status = "rejected"
	srv.repo.saveIncoming(offer)
	go notifyListeners("<", REJECT, offer.Status, srv.incomingStatus)
	return
}

// ================================ Handler ================================

func (srv *service) handleServiceReject() {
	// Register port
	port := srv.register(REJECT)
	for request := range port.Next() {
		go func(request astral.Request) {
			// Accept incoming connection
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", REJECT, "Cannot accept connection from", request.Caller(), err)
				return
			}
			defer conn.Close()
			// Read id of rejected outgoing files
			offerId, err := enc.ReadL8String(conn)
			if err != nil {
				log.Println(">", REJECT, "Cannot read request id", request.Caller(), err)
				return
			}
			// Reject outgoing files
			offer := srv.outgoing[OfferId(offerId)]
			if offer != nil {
				offer.Status.Status = "rejected"
				srv.repo.saveOutgoing(offer)
				go notifyListeners(">", REJECT, offer.Status, srv.outgoingStatus)
			}
			// Send OK
			err = enc.Write(conn, uint8(0))
			if err != nil {
				log.Println(">", REJECT, "Cannot send ok", request.Caller(), err)
				return
			}
		}(*request)
	}
}

// =========================================================================
// ================================ Utils ================================

func (srv *service) register(query string) (port *astral.Port) {
	port, err := astral.Reqister(query)
	if err != nil {
		log.Panicln("Cannot register port", query, err)
	}
	go func() {
		<-srv.ctx.Done()
		_ = port.Close()
	}()
	return
}

func (srv *service) peerMod(peerId PeerId) string {
	peer := srv.peers[peerId]
	if peer == nil {
		peer = &Peer{
			Id:    peerId,
			Alias: "",
			Mod:   "",
		}
		srv.peers[peerId] = peer
	}
	return peer.Mod
}

func notifyListeners(prefix string, query string, data interface{}, listeners []io.WriteCloser) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println(prefix, query, "Cannot read file request", err)
		return
	}
	for _, listener := range listeners {
		_, err := listener.Write(jsonData)
		if err != nil {
			log.Println(prefix, query, "Error while sending files to listener", err)
			return
		}
	}
}
