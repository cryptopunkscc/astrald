package warpdrive

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

func (c LocalClient) Send(peerId PeerId, filePath string) (id OfferId, accepted bool, err error) {
	// Request send
	err = c.Encode("c", localSend)
	if err != nil {
		err = Error(err, "Cannot request send")
		return
	}
	// Send recipient id
	err = c.Encode("[c]c", peerId)
	if err != nil {
		err = Error(err, "Cannot send recipient id")
		return
	}
	// Send file path
	err = c.Encode("[c]c", filePath)
	if err != nil {
		err = Error(err, "Cannot send file path")
		return
	}
	// Read offer id
	err = c.Decode("[c]c", &id)
	if err != nil {
		err = Error(err, "Cannot read offer id")
		return
	}
	// Read result code
	var code byte
	err = c.Decode("c", &code)
	if err != nil {
		err = Error(err, "Cannot read offer result code")
	}
	accepted = code == 1
	return
}

func (c LocalClient) Accept(id OfferId) (err error) {
	// Request accept
	err = c.Encode("c", localAccept)
	if err != nil {
		err = Error(err, "Cannot request accept")
		return
	}

	// Send accepted request id to service
	err = c.Encode("[c]c", id)
	if err != nil {
		err = Error(err, "Cannot send accepted request id")
		return
	}
	// Read OK
	var code byte
	err = c.Decode("c", &code)
	if err != nil {
		err = Error(err, "Cannot read ok")
		return
	}
	return
}

func (c LocalClient) Offers(filter Filter) (offers []Offer, err error) {
	// Request offers
	err = c.Encode("c", localOffers)
	if err != nil {
		err = Error(err, "Cannot request send")
		return
	}

	// Send filter
	if err = cslq.Encode(c.Conn, "[c]c", filter); err != nil {
		err = Error(err, "Cannot send filter")
		return
	}
	// Receive outgoing offers
	if err = json.NewDecoder(c.Conn).Decode(&offers); err != nil {
		err = Error(err, "Cannot read outgoing offers")
		return
	}
	// Send OK
	if err = cslq.Encode(c.Conn, "c", 0); err != nil {
		err = Error(err, "Cannot send ok")
		return
	}
	return
}

func (c LocalClient) Peers() (peers []Peer, err error) {
	// Request peers
	err = c.Encode("c", localPeers)
	if err != nil {
		err = Error(err, "Cannot request peers")
		return
	}
	// Read peers
	if err = json.NewDecoder(c.Conn).Decode(&peers); err != nil {
		err = Error(err, "Cannot read peers")
		return
	}
	// Send OK
	if err = cslq.Encode(c.Conn, "c", 0); err != nil {
		err = Error(err, "Cannot send ok")
		return
	}
	return
}

func (c LocalClient) Status(filter Filter) (status <-chan OfferStatus, err error) {
	// Request status
	err = c.Encode("c", localStatus)
	if err != nil {
		err = Error(err, "Cannot request status")
		return
	}
	// Connect send filter
	err = cslq.Encode(c.Conn, "[c]c", filter)
	if err != nil {
		err = Error(err, "Cannot send filter")
		return
	}
	statChan := make(chan OfferStatus)
	status = statChan
	go func(conn io.ReadWriteCloser, status chan OfferStatus) {
		defer close(status)
		dec := json.NewDecoder(conn)
		files := &OfferStatus{}
		c.Println("Start listening status")
		for {
			err := dec.Decode(files)
			if err != nil {
				if fmt.Sprint(errors.Unwrap(err)) == "use of closed network connection" {
					err = nil
				}
				c.Println(Error(err, "Finish listening offer status"))
				return
			}
			status <- *files
		}
	}(c.Conn, statChan)
	return
}

func (c LocalClient) Subscribe(filter Filter) (offers <-chan Offer, err error) {
	// Request subscribe
	err = c.Encode("c", localSubscribe)
	if err != nil {
		err = Error(err, "Cannot request subscribe")
		return
	}
	err = cslq.Encode(c.Conn, "[c]c", filter)
	if err != nil {
		err = Error(err, "Cannot send filter")
		return
	}
	ofs := make(chan Offer)
	offers = ofs
	go func(conn io.ReadWriteCloser, offers chan Offer) {
		defer close(offers)
		dec := json.NewDecoder(conn)
		files := &Offer{}
		c.Println("Start listening offers")
		for {
			err := dec.Decode(files)
			if err != nil {
				if fmt.Sprint(errors.Unwrap(err)) == "use of closed network connection" {
					err = nil
				}
				c.Println(Error(err, "Finish listening new offers"))
				return
			}
			offers <- *files
		}
	}(c.Conn, ofs)
	return
}

func (c LocalClient) Update(
	peerId PeerId,
	attr string,
	value string,
) (err error) {
	// Request peer update
	err = c.Encode("c", localUpdate)
	if err != nil {
		err = Error(err, "Cannot update")
		return
	}
	// Send peers to update
	req := []string{string(peerId), attr, value}
	err = json.NewEncoder(c.Conn).Encode(req)
	if err != nil {
		err = Error(err, "Cannot send peer update")
		return
	}
	// Wait for OK
	var code byte
	err = cslq.Decode(c.Conn, "c", &code)
	if err != nil {
		err = Error(err, "Cannot read ok")
		return
	}
	return
}
