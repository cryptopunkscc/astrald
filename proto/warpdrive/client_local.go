package warpdrive

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func (c Client) CreateOffer(peerId PeerId, filePath string) (id OfferId, accepted bool, err error) {
	// Request send
	err = c.cslq.Encode("c", localCreateOffer)
	if err != nil {
		err = Error(err, "Cannot request send")
		return
	}
	// Send recipient id
	err = c.cslq.Encode("[c]c", peerId)
	if err != nil {
		err = Error(err, "Cannot send recipient id")
		return
	}
	// Send file path
	err = c.cslq.Encode("[c]c", filePath)
	if err != nil {
		err = Error(err, "Cannot send file path")
		return
	}
	// Read offer id
	err = c.cslq.Decode("[c]c", &id)
	if err != nil {
		err = Error(err, "Cannot read offer id")
		return
	}
	// Read result code
	var code byte
	err = c.cslq.Decode("c", &code)
	if err != nil {
		err = Error(err, "Cannot read offer result code")
	}
	accepted = code == 1
	return
}

func (c Client) AcceptOffer(id OfferId) (err error) {
	// Request accept
	err = c.cslq.Encode("c", localAcceptOffer)
	if err != nil {
		err = Error(err, "Cannot request accept")
		return
	}

	// Send accepted request id to service
	err = c.cslq.Encode("[c]c", id)
	if err != nil {
		err = Error(err, "Cannot send accepted request id")
		return
	}
	// Read OK
	var code byte
	err = c.cslq.Decode("c", &code)
	if err != nil {
		err = Error(err, "Cannot read ok")
		return
	}
	return
}

func (c Client) ListOffers(filter Filter) (offers []Offer, err error) {
	// Request offers
	err = c.cslq.Encode("c", localListOffers)
	if err != nil {
		err = Error(err, "Cannot request send")
		return
	}

	// Send filter
	if err = c.cslq.Encode("c", filter); err != nil {
		err = Error(err, "Cannot send filter")
		return
	}
	// Receive outgoing offers
	if err = json.NewDecoder(c.conn).Decode(&offers); err != nil {
		err = Error(err, "Cannot read outgoing offers")
		return
	}
	return
}

func (c Client) ListPeers() (peers []Peer, err error) {
	// Request peers
	err = c.cslq.Encode("c", localListPeers)
	if err != nil {
		err = Error(err, "Cannot request peers")
		return
	}
	// Read peers
	if err = json.NewDecoder(c.conn).Decode(&peers); err != nil {
		err = Error(err, "Cannot read peers")
		return
	}
	return
}

func (c Client) ListenStatus(filter Filter) (status <-chan OfferStatus, err error) {
	// Request status
	err = c.cslq.Encode("c", localListenStatus)
	if err != nil {
		err = Error(err, "Cannot request status")
		return
	}
	// Connect send filter
	err = c.cslq.Encode("c", filter)
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
		c.log.Println("Start listening status")
		for {
			err := dec.Decode(files)
			if err != nil {
				if fmt.Sprint(errors.Unwrap(err)) == "use of closed network connection" {
					err = nil
				}
				c.log.Println(Error(err, "Finish listening offer status"))
				return
			}
			status <- *files
		}
	}(c.conn, statChan)
	return
}

func (c Client) ListenOffers(filter Filter) (out <-chan Offer, err error) {
	// Request subscribe
	err = c.cslq.Encode("c", localListenOffers)
	if err != nil {
		err = Error(err, "Cannot request subscribe")
		return
	}
	err = c.cslq.Encode("c", filter)
	if err != nil {
		err = Error(err, "Cannot send filter")
		return
	}
	offers := make(chan Offer)
	out = offers
	go func() {
		defer close(offers)
		c.log.Println("Start listening offers")
		dec := json.NewDecoder(c.conn)
		for {
			offer := &Offer{}
			err := dec.Decode(offer)
			if err != nil {
				if fmt.Sprint(errors.Unwrap(err)) == "use of closed network connection" {
					err = nil
				}
				c.log.Print(Error(err, "Finish listening new offers"))
				return
			}
			offers <- *offer
		}
	}()
	return
}

func (c Client) UpdatePeer(
	peerId PeerId,
	attr string,
	value string,
) (err error) {
	// Request peer update
	err = c.cslq.Encode("c", localUpdatePeer)
	if err != nil {
		err = Error(err, "Cannot update")
		return
	}
	// Send peers to update
	req := []string{string(peerId), attr, value}
	err = json.NewEncoder(c.conn).Encode(req)
	if err != nil {
		err = Error(err, "Cannot send peer update")
		return
	}
	// Wait for OK
	var code byte
	err = c.cslq.Decode("c", &code)
	if err != nil {
		err = Error(err, "Cannot read ok")
		return
	}
	return
}
