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
	go service.handleSenderSendFile()
	go service.handleSenderSendingStatus()
	go service.handleSenderSentRequests()
	go service.handleSenderEventsSubscribe()
	go service.handleRecipientAcceptRequest()
	go service.handleRecipientRejectRequest()
	go service.handleRecipientIncomingFiles()
	go service.handleRecipientReceivedRequests()
	go service.handleRecipientUpdatePeer()
	go service.handleRecipientEventsSubscribe()
	go service.handleCommandLine()
	<-ctx.Done()
}

func newService(ctx context.Context) *service {
	service := &service{
		ctx:      ctx,
		home:     userFiles(),
		received: receivedFiles(),
		peers:    map[string]*Peer{},
		requests: requests{
			incoming: map[OfferId]*Offer{},
			outgoing: map[OfferId]*Offer{},
		},
	}
	service.setupIdentity()
	service.setupPeers()
	return service
}

type service struct {
	ctx      context.Context
	identity string
	home     storage
	received storage
	peers    map[string]*Peer
	requests
	notify
}

type requests struct {
	incoming map[OfferId]*Offer
	outgoing map[OfferId]*Offer
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
		srv.peers[contact.Id] = &Peer{
			Id:    PeerId(contact.Id),
			Alias: contact.Name,
		}
	}
}

// =========================================================================
// ================================ Caller =================================

func (srv *service) sendFile(peer string, path string) (id string, err error) {
	// Get files info
	files, err := srv.home.Info(path)
	if err != nil {
		log.Println("<", SEND, "Cannot get files info", peer, err)
		return
	}
	// Connect to service
	conn, err := astral.Query(peer, SEND)
	if err != nil {
		log.Println("<", SEND, "Cannot connect", peer, err)
		return
	}
	// Send file request
	id = newOfferId()
	err = enc.WriteL8String(conn, id)
	if err != nil {
		log.Println("<", SEND, "Cannot send offer id", peer, err)
		return "", err
	}
	shrunken := ShrinkPaths(files)
	err = json.NewEncoder(conn).Encode(shrunken)
	if err != nil {
		log.Println("<", SEND, "Cannot send offer info", peer, id, err)
		return
	}
	// Wait for close
	_, err = enc.ReadUint8(conn)
	if err != nil {
		log.Println("<", SEND, "Rejected", peer, err)
		return
	}
	// Cache outgoing files request
	status := Status{
		Id:     OfferId(id),
		Status: "sent",
	}
	srv.outgoing[status.Id] = &Offer{
		Status: status,
		Files:  files,
	}
	// Notify status listeners
	go notifyListeners("<", SEND, status, srv.outgoingStatus)
	return
}

func newOfferId() string {
	v4, err := uuid.NewV4()
	if err != nil {
		log.Panic(err)
	}
	return v4.String()
}

// ================================ Handler ================================

func (srv *service) handleServiceSend() {
	// Register port
	port := srv.register(SEND)
	for request := range port.Next() {
		// Handle received request
		go func(request *astral.Request) {
			caller := request.Caller()
			peerMode := srv.peerMod(caller)
			// Check if peer is blocked
			if peerMode == PEER_MOD_BLOCK {
				request.Reject()
				log.Println(">", SEND, "Blocked request from", caller)
				return
			}
			// Accept connection
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", SEND, "Cannot accept request from", caller, err)
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
				Peer:  PeerId(caller),
				Status: Status{
					Id:     OfferId(offerId),
					Status: "received",
				},
			}
			srv.incoming[offer.Id] = offer
			// Send OK
			_ = enc.Write(conn, uint8(0))
			// Notify status listeners
			go notifyListeners(">", SEND, offer.Status, srv.incomingStatus)
			// Auto accept incoming offer if peer is trusted
			switch peerMode {
			case PEER_MOD_TRUST:
				_ = srv.acceptIncomingFiles(offer.Id)
			case PEER_MOD_ASK:
				go notifyListeners(">", SEND, offer, srv.filesRequest)
			}
		}(request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (srv *service) acceptIncomingFiles(id OfferId) (err error) {
	// Get cached incoming files by request id
	files := srv.incoming[id]
	if files == nil {
		err = errors.New(fmt.Sprint("No incoming file with id ", id))
		log.Println("<", ACCEPT, "Cannot find incoming file", err)
		return err
	}
	// Obtain files reader connection
	filesConn, err := func() (filesConn io.ReadWriteCloser, err error) {
		// Connect to service
		conn, err := astral.Query(string(files.Peer), ACCEPT)
		if err != nil {
			log.Println("<", ACCEPT, "Cannot connect", err)
			return
		}
		defer conn.Close()
		// Send file request id
		err = enc.WriteL8String(conn, string(files.Id))
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
		files.Status.Status = "accepted"
		go notifyListeners("<", ACCEPT, files.Status, srv.incomingStatus)
		err = enc.Write(conn, uint8(0))
		if err != nil {
			log.Println("<", ACCEPT, "Cannot send ok", err)
			return
		}
		// Open connection for downloading files
		filesConn, err = astral.Query(string(files.Peer), filesQuery)
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
		for _, file := range files.Files {
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
							files.Status.Status = fmt.Sprintf("download: %s %d/%dB", file.Path, progress, size)
							go notifyListeners("<", ACCEPT, files.Status, srv.incomingStatus)
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
		files.Status.Status = "downloaded"
		go notifyListeners("<", ACCEPT, files.Status, srv.incomingStatus)
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
		go func(request *astral.Request) {
			// Accept incoming connection
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", ACCEPT, "Cannot accept connection from", request.Caller(), err)
				return
			}
			defer conn.Close()
			// Read request id
			id, err := enc.ReadL8String(conn)
			if err != nil {
				log.Println(">", ACCEPT, "Cannot read file request id", err)
				return
			}
			// Obtain file by request id
			files := srv.outgoing[OfferId(id)]
			if files == nil {
				log.Println(">", ACCEPT, "Cannot read file request id", err)
				conn.Close()
			}
			files.Status.Status = "accepted"
			go notifyListeners(">", ACCEPT, files.Status, srv.outgoingStatus)
			// Register port for reading files
			filesQuery := PORT + "/" + string(files.Id)
			filesPort, err := astral.Reqister(filesQuery)
			if err != nil {
				log.Println(">", ACCEPT, "Cannot register port for", filesPort, err)
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
				log.Println(">", ACCEPT, "Rejected by peer", filesQuery, err)
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
			// Send files
			for _, file := range files.Files {
				if !file.IsDir {
					reader, err := srv.home.Reader(file.Path)
					if err != nil {
						return
					}
					progress := &ioprogress.Reader{
						Reader:       reader,
						Size:         file.Size,
						DrawInterval: 200 * time.Millisecond,
						DrawFunc: func(progress int64, size int64) error {
							files.Status.Status = fmt.Sprintf("upload %s %d/%dB", file.Path, progress, size)
							go notifyListeners(">", ACCEPT, files.Status, srv.outgoingStatus)
							return nil
						},
					}
					_, err = io.CopyN(filesConn, progress, file.Size)
					if err != nil {
						return
					}
				}
			}
			// Wait OK
			_, err = enc.ReadUint8(filesConn)
			if err != nil {
				return
			}
			files.Status.Status = "uploaded"
			go notifyListeners(">", ACCEPT, files.Status, srv.outgoingStatus)
		}(request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (srv *service) rejectIncomingFiles(id OfferId) (err error) {
	// Get cached incoming files by id
	files := srv.incoming[id]
	if files == nil {
		err = errors.New(fmt.Sprint("No incoming file with id", id))
		log.Println("<", REJECT, "Cannot find incoming file", err)
		return
	}
	// Connect to service
	conn, err := astral.Query(string(files.Peer), REJECT)
	if err != nil {
		log.Println("<", REJECT, "Cannot connect", err)
		return
	}
	defer conn.Close()
	// Send rejected files id
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		log.Println("<", REJECT, "Write rejected request id", id, err)
		return
	}
	// Wait for OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		log.Println("<", REJECT, "Rejected request id", id, err)
		return
	}
	files.Status.Status = "rejected"
	go notifyListeners("<", REJECT, files.Status, srv.incomingStatus)
	return
}

// ================================ Handler ================================

func (srv *service) handleServiceReject() {
	// Register port
	port := srv.register(REJECT)
	for request := range port.Next() {
		go func(request *astral.Request) {
			// Accept incoming connection
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", REJECT, "Cannot accept connection from", request.Caller(), err)
				return
			}
			defer conn.Close()
			// Read id of rejected outgoing files
			requestId, err := enc.ReadL8String(conn)
			if err != nil {
				log.Println(">", REJECT, "Cannot read request id", request.Caller(), err)
				return
			}
			// Reject outgoing files
			req := srv.outgoing[OfferId(requestId)]
			if req != nil {
				req.Status.Status = "rejected"
				go notifyListeners(">", REJECT, req.Status, srv.outgoingStatus)
			}
			// Send OK
			err = enc.Write(conn, uint8(0))
			if err != nil {
				log.Println(">", REJECT, "Cannot send ok", request.Caller(), err)
				return
			}
		}(request)
	}
}

// =========================================================================
// ================================ Utils ================================

func (srv *service) register(query string) (port *astral.Port) {
	port, err := astral.Reqister(query)
	if err != nil {
		log.Panic(err)
	}
	go func() {
		<-srv.ctx.Done()
		_ = port.Close()
	}()
	return
}

func (srv *service) peerMod(caller string) string {
	peer := srv.peers[caller]
	if peer == nil {
		peer = &Peer{
			Id:    PeerId(caller),
			Alias: "",
			Mod:   "",
		}
		srv.peers[caller] = peer
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
