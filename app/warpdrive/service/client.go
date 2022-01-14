package warpdrive

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
	"log"
	"sync"
)

// SENDER
const (
	SEN_PEERS  = PORT + "/sender/peers"
	SEN_SEND   = PORT + "/sender/send"
	SEN_STATUS = PORT + "/sender/status"
	SEN_SENT   = PORT + "/sender/sent"
	SEN_EVENTS = PORT + "/sender/events"
)

// RECIPIENT
const (
	REC_INCOMING = PORT + "/recipient/incoming"
	REC_RECEIVED = PORT + "/recipient/received"
	REC_ACCEPT   = PORT + "/recipient/accept"
	REC_REJECT   = PORT + "/recipient/reject"
	REC_UPDATE   = PORT + "/recipient/update"
	REC_EVENTS   = PORT + "/recipient/events"
)

func NewUIClient() ClientApi {
	return &client{sender{}, recipient{}}
}

type sender struct{}
type recipient struct{}
type client struct {
	sender
	recipient
}

func (s *client) Sender() SenderApi {
	return &s.sender
}

func (s *client) Recipient() RecipientApi {
	return &s.recipient
}

// =========================================================================
// ================================ Caller =================================

func (s *sender) Peers() (peers []Peer, err error) {
	// Connect to local service
	conn, err := astral.Query("", SEN_PEERS)
	if err != nil {
		log.Println("<", SEN_PEERS, "Cannot connect to service", err)
		return
	}
	defer conn.Close()
	// Read peers
	err = json.NewDecoder(conn).Decode(&peers)
	if err != nil {
		log.Println("<", SEN_PEERS, "Cannot read peers", err)
		return
	}
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		log.Println("<", SEN_PEERS, "Cannot send ok", err)
		return
	}
	return
}

// ================================ Handler ================================

func (srv *service) handleSenderPeers() {
	// Register port
	port := srv.register(SEN_PEERS)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", SEN_PEERS, "Cannot accept request", err)
				return
			}
			defer conn.Close()
			// Collect peers
			peers := make([]*Peer, len(srv.peers))
			i := 0
			peersMap := srv.peers
			for key := range peersMap {
				peers[i] = srv.peers[key]
				i++
			}
			// Send peers
			err = json.NewEncoder(conn).Encode(peers)
			if err != nil {
				log.Println(">", SEN_PEERS, "Cannot send peers", err)
				return
			}
			// Read OK
			_, err = enc.ReadUint8(conn)
			if err != nil {
				log.Println(">", SEN_PEERS, "Cannot read ok", err)
				return
			}
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (s *sender) Send(peerId PeerId, filePath string) (id OfferId, err error) {
	// Connect to local service
	conn, err := astral.Query("", SEN_SEND)
	if err != nil {
		log.Println("<", SEN_SEND, "Cannot connect to service", err)
		return
	}
	defer conn.Close()
	// Send recipient id
	err = enc.WriteL8String(conn, string(peerId))
	if err != nil {
		log.Println("<", SEN_SEND, "Cannot send recipient id", err)
		return
	}
	// Send file path
	err = enc.WriteL8String(conn, filePath)
	if err != nil {
		log.Println("<", SEN_SEND, "Cannot send file path", err)
		return
	}
	// Read response
	strId, err := enc.ReadL8String(conn)
	if err != nil {
		log.Println("<", SEN_SEND, "Cannot read response id", err)
	}
	id = OfferId(strId)
	return
}

// ================================ Handler ================================

func (srv *service) handleSenderSend() {
	// Register port
	port := srv.register(SEN_SEND)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", SEN_SEND, "Cannot accept request", err)
				return
			}
			defer conn.Close()
			// Read peer id
			peerId, err := enc.ReadL8String(conn)
			if err != nil {
				log.Println(">", SEN_SEND, "Cannot read peer id", err)
				return
			}
			// Read file path
			filePath, err := enc.ReadL8String(conn)
			if err != nil {
				log.Println(">", SEN_SEND, "Cannot read file path", err)
				return
			}
			// Get files info
			files, err := srv.resolver.Info(filePath)
			if err != nil {
				log.Println("<", SEND, "Cannot get files info", err)
				return
			}
			// Send file to recipient service
			id, err := srv.callServiceSend(peerId, files)
			if err != nil {
				log.Println(">", SEN_SEND, "Cannot send file", err)
				return
			}
			// Write response to sender
			err = enc.WriteL8String(conn, id)
			if err != nil {
				log.Println(">", SEN_SEND, "Cannot send id", id, err)
				return
			}
			log.Println(">", SEN_SEND, filePath, "offer sent to", peerId)
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (s *sender) Status(id OfferId) (status string, err error) {
	// Connect to service
	conn, err := astral.Query("", SEN_STATUS)
	if err != nil {
		log.Println("<", SEN_STATUS, "Cannot connect to service", err)
		return
	}
	defer conn.Close()
	// Send request id
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		log.Println("<", SEN_STATUS, "Cannot send request id", err)
		return
	}
	// Receive status
	status, err = enc.ReadL8String(conn)
	if err != nil {
		log.Println("<", SEN_STATUS, "Cannot read request status", err)
	}
	return
}

// ================================ Handler ================================

func (srv *service) handleSenderStatus() {
	// Register port
	port := srv.register(SEN_STATUS)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", SEN_STATUS, "Cannot accept request", err)
				return
			}
			id, err := enc.ReadL8String(conn)
			if err != nil {
				log.Println(">", SEN_STATUS, "Cannot read request id", err)
				return
			}
			files := srv.outgoing[OfferId(id)]
			if files == nil {
				log.Println(">", SEN_STATUS, "Cannot find outgoing files with id", id)
				return
			}
			err = enc.WriteL8String(conn, files.Status.Status)
			if err != nil {
				log.Println(">", SEN_STATUS, "Cannot send file status", files.Status, err)
				return
			}
			log.Println(">", SEN_STATUS, "Send file status", files.Status, err)
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (s *sender) Sent() (offers Offers, err error) {
	// Connect to service
	conn, err := astral.Query("", SEN_SENT)
	if err != nil {
		log.Println("<", SEN_SENT, "Cannot connect to service", err)
		return
	}
	defer conn.Close()
	// Receive outgoing offers
	err = json.NewDecoder(conn).Decode(&offers)
	if err != nil {
		log.Println("<", SEN_SENT, "Cannot read outgoing offers", err)
		return
	}
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		log.Println("<", SEN_SENT, "Cannot send ok", err)
		return
	}
	return
}

// ================================ Handler ================================

func (srv *service) handleSenderSent() {
	// Register port
	port := srv.register(SEN_SENT)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			// Accept connection
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", SEN_SENT, "Cannot accept request", err)
				return
			}
			defer conn.Close()
			// Send outgoing files
			err = json.NewEncoder(conn).Encode(srv.outgoing)
			if err != nil {
				log.Println(">", SEN_SENT, "Cannot send outgoing offers", err)
				return
			}
			// Wait for OK
			_, err = enc.ReadUint8(conn)
			if err != nil {
				log.Println(">", SEN_SENT, "Cannot read ok", err)
				return
			}
			log.Println(">", SEN_SENT, "Send outgoing offers")
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (s *sender) Events() (outgoing <-chan Status, err error) {
	// Connect to local service
	conn, err := astral.Query("", SEN_EVENTS)
	if err != nil {
		log.Println("<", SEN_EVENTS, "Cannot connect to service", err)
		return
	}
	out := make(chan Status)
	outgoing = out
	go func(conn io.ReadWriteCloser, inc chan Status) {
		defer close(inc)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &Status{}
		log.Println("<", SEN_EVENTS, "Start listening status")
		for true {
			err := dec.Decode(files)
			if err != nil {
				log.Println("<", SEN_EVENTS, "Finish listening offers status", err)
				return
			}
			inc <- *files
		}
	}(conn, out)
	return
}

// ================================ Handler ================================

func (srv *service) handleSenderEvents() {
	// Register port
	port := srv.register(SEN_EVENTS)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", SEN_EVENTS, "Cannot accept request", err)
				return
			}
			defer conn.Close()
			srv.notify.mu.Lock()
			srv.outgoingStatus = append(srv.outgoingStatus, conn)
			srv.notify.mu.Unlock()
			_, _ = enc.ReadUint8(conn)
			// TODO remove listener
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (r *recipient) Offers() (offers <-chan Offer, err error) {
	// Connect to local service
	conn, err := astral.Query("", REC_INCOMING)
	if err != nil {
		log.Println("<", REC_INCOMING, "Cannot connect to service", err)
		return
	}
	ofs := make(chan Offer)
	offers = ofs
	go func(conn io.ReadWriteCloser, offers chan Offer) {
		defer close(offers)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &Offer{}
		log.Println("<", REC_INCOMING, "Start listening offers")
		for true {
			err := dec.Decode(files)
			if err != nil {
				log.Println("<", REC_INCOMING, "Finish listening offers", err)
				return
			}
			offers <- *files
		}
	}(conn, ofs)
	return
}

// ================================ Handler ================================

func (srv *service) handleRecipientOffers() {
	// Register port
	port := srv.register(REC_INCOMING)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", REC_INCOMING, "Cannot accept request", err)
				return
			}
			defer conn.Close()
			srv.mu.Lock()
			srv.filesRequest = append(srv.filesRequest, conn)
			srv.mu.Unlock()
			_, _ = enc.ReadUint8(conn)
			// TODO remove listener
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (r *recipient) Received() (offers Offers, err error) {
	// Connect to service
	conn, err := astral.Query("", REC_RECEIVED)
	if err != nil {
		log.Println("<", REC_RECEIVED, "Cannot connect to service", err)
		return
	}
	defer conn.Close()
	// Receive outgoing offers
	err = json.NewDecoder(conn).Decode(&offers)
	if err != nil {
		log.Println("<", REC_RECEIVED, "Cannot read outgoing offers", err)
		return
	}
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		log.Println("<", REC_RECEIVED, "Cannot send ok", err)
		return
	}
	return
}

// ================================ Handler ================================

func (srv *service) handleRecipientReceived() {
	// Register port
	port := srv.register(REC_RECEIVED)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			// Accept connection
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", REC_RECEIVED, "Cannot accept request", err)
				return
			}
			defer conn.Close()
			// Send outgoing files
			err = json.NewEncoder(conn).Encode(srv.incoming)
			if err != nil {
				log.Println(">", REC_RECEIVED, "Cannot send incoming offers", err)
				return
			}
			// Wait for OK
			_, err = enc.ReadUint8(conn)
			if err != nil {
				log.Println(">", REC_RECEIVED, "Cannot read ok", err)
				return
			}
			log.Println(">", REC_RECEIVED, "Send incoming offers")
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (r *recipient) Accept(id OfferId) (err error) {
	// Connect to local service
	conn, err := astral.Query("", REC_ACCEPT)
	if err != nil {
		log.Println("<", REC_ACCEPT, "Cannot connect to service", err)
		return
	}
	defer conn.Close()
	// Send accepted request id to service
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		log.Println("<", REC_ACCEPT, "Cannot send accepted request id", err)
		return
	}
	// Read OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		log.Println("<", REC_ACCEPT, "Cannot read ok", err)
		return
	}
	return
}

// ================================ Handler ================================

func (srv *service) handleRecipientAccept() {
	// Register port
	port := srv.register(REC_ACCEPT)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", REC_ACCEPT, "Cannot accept request", err)
				return
			}
			defer conn.Close()
			id, err := enc.ReadL8String(conn)
			if err != nil {
				log.Println(">", REC_ACCEPT, "Cannot read request id", err)
				return
			}
			err = srv.callServiceAccept(OfferId(id))
			if err != nil {
				log.Println(">", REC_ACCEPT, "Cannot accept incoming files", id, err)
				return
			}
			err = enc.Write(conn, uint8(0))
			if err != nil {
				log.Println(">", REC_ACCEPT, "Cannot send ok", err)
				return
			}
			log.Println(">", REC_ACCEPT, "Accepted incoming files", id)
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (r *recipient) Reject(id OfferId) (err error) {
	// Connect to local service
	conn, err := astral.Query("", REC_REJECT)
	if err != nil {
		log.Println("<", REC_REJECT, "Cannot connect to service", err)
		return
	}
	// Send accepted request id to service
	defer conn.Close()
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		log.Println("<", REC_REJECT, "Cannot send rejected request id", err)
		return
	}
	// Read OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		log.Println("<", REC_REJECT, "Cannot read ok", err)
		return
	}
	return
}

// ================================ Handler ================================

func (srv *service) handleRecipientReject() {
	// Register port
	port := srv.register(REC_REJECT)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", REC_REJECT, "Cannot accept request", err)
				return
			}
			defer conn.Close()
			id, err := enc.ReadL8String(conn)
			if err != nil {
				log.Println(">", REC_REJECT, "Cannot read request id", err)
				return
			}
			err = srv.callServiceReject(OfferId(id))
			if err != nil {
				log.Println(">", REC_REJECT, "Cannot reject incoming files", id, err)
				return
			}
			err = enc.Write(conn, uint8(0))
			if err != nil {
				log.Println(">", REC_ACCEPT, "Cannot send ok", err)
				return
			}
			log.Println(">", REC_REJECT, "Rejected incoming files", id, err)
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (r *recipient) Update(
	peerId PeerId,
	attr string,
	value string,
) (err error) {
	// Connect to local service
	conn, err := astral.Query("", REC_UPDATE)
	if err != nil {
		log.Println("<", REC_UPDATE, "Cannot connect to service", err)
		return
	}
	defer conn.Close()
	// Send peers to update
	req := []string{string(peerId), attr, value}
	err = json.NewEncoder(conn).Encode(req)
	if err != nil {
		log.Println("<", REC_UPDATE, "Cannot send peer update", err)
		return
	}
	// Wait for OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		log.Println("<", REC_UPDATE, "Cannot read ok", err)
		return
	}
	return
}

// ================================ Handler ================================

func (srv *service) handleRecipientUpdate() {
	// Register port
	port := srv.register(REC_UPDATE)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			// Accept connection
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", REC_UPDATE, "Cannot accept request", err)
				return
			}
			defer conn.Close()
			// Read peer update
			var req []string
			err = json.NewDecoder(conn).Decode(&req)
			if err != nil {
				log.Println(">", REC_UPDATE, "Cannot read peer update", err)
				return
			}
			peerId := req[0]
			attr := req[1]
			value := req[2]
			peer := srv.peers[PeerId(peerId)]
			if peer == nil {
				log.Println(">", REC_UPDATE, "Cannot find peer with id", peerId)
				return
			}
			switch attr {
			case "mod":
				peer.Mod = value
			case "alias":
				peer.Alias = value
			default:
				log.Println(">", REC_UPDATE, "Invalid peer attribute", attr)
				return
			}
			var peers []Peer
			for _, p := range srv.peers {
				peers = append(peers, *p)
			}
			srv.repo.savePeers(peers)
			// Send OK
			err = enc.Write(conn, uint8(0))
			if err != nil {
				log.Println(">", REC_UPDATE, "Cannot send ok", err)
				return
			}
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (r *recipient) Events() (incoming <-chan Status, err error) {
	// Connect to local service
	conn, err := astral.Query("", REC_EVENTS)
	if err != nil {
		log.Println(REC_EVENTS, "Cannot connect to service", err)
		return
	}
	inc := make(chan Status)
	incoming = inc
	go func(conn io.ReadWriteCloser, inc chan Status) {
		defer close(inc)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &Status{}
		for true {
			err := dec.Decode(files)
			if err != nil {
				log.Println(REC_EVENTS, "Cannot decode status", err)
				return
			}
			inc <- *files
		}
	}(conn, inc)
	return
}

// ================================ Handler ================================

func (srv *service) handleRecipientEvents() {
	// Register port
	port := srv.register(REC_EVENTS)
	for request := range port.Next() {
		// Handle received request
		go func(request astral.Request) {
			if srv.isRejected(&request) {
				return
			}
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", REC_EVENTS, "Cannot accept request", err)
				return
			}
			defer conn.Close()
			srv.notify.mu.Lock()
			srv.incomingStatus = append(srv.incomingStatus, conn)
			srv.notify.mu.Unlock()
			_, _ = enc.ReadUint8(conn)
			// TODO remove listener
		}(*request)
	}
}

// =========================================================================
// ================================ Caller =================================

func (s *client) Events() (events <-chan Status, err error) {
	senderEvents, err := s.sender.Events()
	if err != nil {
		return
	}
	recipientEvents, err := s.recipient.Events()
	if err != nil {
		return
	}
	events = merge(senderEvents, recipientEvents)
	return
}

func merge(cs ...<-chan Status) <-chan Status {
	out := make(chan Status)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan Status) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// =========================================================================
// ================================ Utils ==================================

func (srv *service) isRejected(request *astral.Request) bool {
	caller := request.Caller()
	isRemote := caller != "" && caller != srv.identity
	if isRemote {
		request.Reject()
		log.Println(">", request.Query(), "Accept only local request, by caller was remote", caller)
	}
	return isRemote
}
