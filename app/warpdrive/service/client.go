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

func NewClient() ClientApi {
	c := client{log.Default()}
	return &clientApi{
		sender{c},
		recipient{c},
	}
}

type clientApi struct {
	sender
	recipient
}
type sender struct {
	client
}

type recipient struct {
	client
}

type client struct {
	*log.Logger
}

func (s *clientApi) Sender() SenderApi {
	return &s.sender
}

func (s *clientApi) Recipient() RecipientApi {
	return &s.recipient
}

// =========================================================================
// ================================ Caller =================================

func (s sender) Peers() (peers []Peer, err error) {
	// Connect to local service
	conn, err := s.query(SEN_PEERS)
	if err != nil {
		return
	}
	defer conn.Close()
	// Read peers
	err = json.NewDecoder(conn).Decode(&peers)
	if err != nil {
		s.Println("Cannot read peers", err)
		return
	}
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		s.Println("Cannot send ok", err)
		return
	}
	return
}

// ================================ Handler ================================

func handleSenderPeers(srv service, request astral.Request) {
	if srv.isRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	// Get peers
	peers := srv.listPeers()
	// Send peers
	err = json.NewEncoder(conn).Encode(peers)
	if err != nil {
		srv.Println("Cannot send peers", err)
		return
	}
	// Read OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		srv.Println("Cannot read ok", err)
		return
	}
}

// =========================================================================
// ================================ Caller =================================

func (s sender) Send(peerId PeerId, filePath string) (id OfferId, err error) {
	// Connect to local service
	conn, err := s.query(SEN_SEND)
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
	id = OfferId(strId)
	return
}

// ================================ Handler ================================

func handleSenderSend(srv service, request astral.Request) {
	if srv.isRejected(request) {
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
	files, err := srv.resolver.Info(filePath)
	if err != nil {
		srv.Println("Cannot get files info", err)
		return
	}
	// Send file to recipient service
	id, err := srv.callSend(peerId, files)
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

// =========================================================================
// ================================ Caller =================================

func (s sender) Status(id OfferId) (status string, err error) {
	// Connect to service
	conn, err := s.query(SEN_STATUS)
	if err != nil {
		return
	}
	defer conn.Close()
	// Send request id
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		s.Println("Cannot send request id", err)
		return
	}
	// Receive status
	status, err = enc.ReadL8String(conn)
	if err != nil {
		s.Println("Cannot read request status", err)
	}
	return
}

// ================================ Handler ================================

func handleSenderStatus(srv service, request astral.Request) {
	if srv.isRejected(request) {
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
	files := srv.getOutgoingOffer(OfferId(id))
	if files == nil {
		srv.Println("Cannot find outgoing files with id", id)
		return
	}
	err = enc.WriteL8String(conn, files.Status.Status)
	if err != nil {
		srv.Println("Cannot send file status", files.Status, err)
		return
	}
	srv.Println("Send file status", files.Status, err)
}

// =========================================================================
// ================================ Caller =================================

func (s sender) Sent() (offers Offers, err error) {
	// Connect to service
	conn, err := s.query(SEN_SENT)
	if err != nil {
		return
	}
	defer conn.Close()
	// Receive outgoing offers
	err = json.NewDecoder(conn).Decode(&offers)
	if err != nil {
		s.Println("Cannot read outgoing offers", err)
		return
	}
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		s.Println("Cannot send ok", err)
		return
	}
	return
}

// ================================ Handler ================================

func handleSenderSent(srv service, request astral.Request) {
	if srv.isRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	// Send outgoing files
	err = json.NewEncoder(conn).Encode(srv.outgoing)
	if err != nil {
		srv.Println("Cannot send outgoing offers", err)
		return
	}
	// Wait for OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		srv.Println("Cannot read ok", err)
		return
	}
	srv.Println("Send outgoing offers")
}

// =========================================================================
// ================================ Caller =================================

func (s sender) Events() (outgoing <-chan Status, err error) {
	// Connect to local service
	conn, err := s.query(SEN_EVENTS)
	if err != nil {
		return
	}
	out := make(chan Status)
	outgoing = out
	go func(conn io.ReadWriteCloser, inc chan Status) {
		defer close(inc)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &Status{}
		s.Println("Start listening status")
		for {
			err := dec.Decode(files)
			if err != nil {
				s.Println("Finish listening offers status", err)
				return
			}
			inc <- *files
		}
	}(conn, out)
	return
}

// ================================ Handler ================================

func handleSenderEvents(srv service, request astral.Request) {
	if srv.isRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	remove := srv.outgoingStatus.subscribe(conn)
	defer remove()
	// Wait for close
	_, _ = enc.ReadUint8(conn)
}

// =========================================================================
// ================================ Caller =================================

func (r recipient) Offers() (offers <-chan Offer, err error) {
	// Connect to local service
	conn, err := r.query(REC_INCOMING)
	if err != nil {
		return
	}
	ofs := make(chan Offer)
	offers = ofs
	go func(conn io.ReadWriteCloser, offers chan Offer) {
		defer close(offers)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &Offer{}
		r.Println("Start listening offers")
		for {
			err := dec.Decode(files)
			if err != nil {
				r.Println("Finish listening offers", err)
				return
			}
			offers <- *files
		}
	}(conn, ofs)
	return
}

// ================================ Handler ================================

func handleRecipientOffers(srv service, request astral.Request) {
	if srv.isRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	remove := srv.filesOffers.subscribe(conn)
	defer remove()
	// Wait for close
	_, _ = enc.ReadUint8(conn)
}

// =========================================================================
// ================================ Caller =================================

func (r recipient) Received() (offers Offers, err error) {
	// Connect to service
	conn, err := r.query(REC_RECEIVED)
	if err != nil {
		return
	}
	defer conn.Close()
	// Receive outgoing offers
	err = json.NewDecoder(conn).Decode(&offers)
	if err != nil {
		r.Println("Cannot read outgoing offers", err)
		return
	}
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		r.Println("Cannot send ok", err)
		return
	}
	return
}

// ================================ Handler ================================

func handleRecipientReceived(srv service, request astral.Request) {
	if srv.isRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	// Send outgoing files
	err = json.NewEncoder(conn).Encode(srv.incoming)
	if err != nil {
		srv.Println("Cannot send incoming offers", err)
		return
	}
	// Wait for OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		srv.Println("Cannot read ok", err)
		return
	}
	srv.Println("Send incoming offers")
}

// =========================================================================
// ================================ Caller =================================

func (r recipient) Accept(id OfferId) (err error) {
	// Connect to local service
	conn, err := r.query(REC_ACCEPT)
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

// ================================ Handler ================================

func handleRecipientAccept(srv service, request astral.Request) {
	if srv.isRejected(request) {
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
	err = srv.callAccept(OfferId(id))
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

// =========================================================================
// ================================ Caller =================================

func (r recipient) Reject(id OfferId) (err error) {
	// Connect to local service
	conn, err := r.query(REC_REJECT)
	if err != nil {
		return
	}
	// Send accepted request id to service
	defer conn.Close()
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		r.Println("Cannot send rejected request id", err)
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

// ================================ Handler ================================

func handleRecipientReject(srv service, request astral.Request) {
	if srv.isRejected(request) {
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
	err = srv.callReject(OfferId(id))
	if err != nil {
		srv.Println("Cannot reject incoming files", id, err)
		return
	}
	err = enc.Write(conn, uint8(0))
	if err != nil {
		srv.Println("Cannot send ok", err)
		return
	}
	srv.Println("Rejected incoming files", id, err)
}

// =========================================================================
// ================================ Caller =================================

func (r recipient) Update(
	peerId PeerId,
	attr string,
	value string,
) (err error) {
	// Connect to local service
	conn, err := r.query(REC_UPDATE)
	if err != nil {
		return
	}
	defer conn.Close()
	// Send peers to update
	req := []string{string(peerId), attr, value}
	err = json.NewEncoder(conn).Encode(req)
	if err != nil {
		r.Println("Cannot send peer update", err)
		return
	}
	// Wait for OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		r.Println("Cannot read ok", err)
		return
	}
	return
}

// ================================ Handler ================================

func handleRecipientUpdate(srv service, request astral.Request) {
	if srv.isRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	// Read peer update
	var req []string
	err = json.NewDecoder(conn).Decode(&req)
	if err != nil {
		srv.Println("Cannot read peer update", err)
		return
	}
	peerId := req[0]
	attr := req[1]
	value := req[2]
	// Update peer
	srv.updatePeer(peerId, attr, value)
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		srv.Println("Cannot send ok", err)
		return
	}
}

// =========================================================================
// ================================ Caller =================================

func (r recipient) Events() (incoming <-chan Status, err error) {
	// Connect to local service
	conn, err := r.query(REC_EVENTS)
	if err != nil {
		return
	}
	inc := make(chan Status)
	incoming = inc
	go func(conn io.ReadWriteCloser, inc chan Status) {
		defer close(inc)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &Status{}
		r.Println("Start listening status")
		for {
			err := dec.Decode(files)
			if err != nil {
				r.Println("Cannot decode status", err)
				return
			}
			inc <- *files
		}
	}(conn, inc)
	return
}

// ================================ Handler ================================

func handleRecipientEvents(srv service, request astral.Request) {
	if srv.isRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	remove := srv.incomingStatus.subscribe(conn)
	defer remove()
	// Wait for close
	_, _ = enc.ReadUint8(conn)
}

// =========================================================================
// ================================ Caller =================================

func (s *clientApi) Events() (events <-chan Status, err error) {
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

func (client *client) query(port string) (conn io.ReadWriteCloser, err error) {
	client.Logger = newLogger("<", port)
	// Connect to local service
	conn, err = astral.Query("", port)
	if err != nil {
		client.Println("Cannot connect to service", err)
	}
	return
}

func (srv *service) isRejected(request astral.Request) bool {
	caller := request.Caller()
	isRemote := caller != "" && caller != srv.identity
	if isRemote {
		request.Reject()
		srv.Println("Accept only local request, but caller was remote", caller)
	}
	return isRemote
}
