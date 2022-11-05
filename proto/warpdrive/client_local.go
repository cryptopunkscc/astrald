package warpdrive

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func (c Client) CreateOffer(peerId PeerId, filePath string) (id OfferId, accepted bool, err error) {
	// Request create offer
	err = c.cslq.Encode("c [c]c [c]c", localCreateOffer, peerId, filePath)
	if err != nil {
		err = Error(err, "Cannot create offer")
		return
	}
	// Read result
	var code byte
	err = c.cslq.Decode("[c]c c", &id, &code)
	if err != nil {
		err = Error(err, "Cannot read create offer results")
		return
	}
	accepted = code == 1
	return
}

func (c Client) AcceptOffer(id OfferId) (err error) {
	// Request accept offer
	err = c.cslq.Encode("c [c]c", localAcceptOffer, id)
	if err != nil {
		err = Error(err, "Cannot request accept")
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
	// Request list offers
	err = c.cslq.Encode("c c", localListOffers, filter)
	if err != nil {
		err = Error(err, "Cannot request offer list")
		return
	}
	// Receive offers
	if err = json.NewDecoder(c.conn).Decode(&offers); err != nil {
		err = Error(err, "Cannot read offers")
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
	err = c.cslq.Encode("c c", localListenStatus, filter)
	if err != nil {
		err = Error(err, "Cannot request status")
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
	err = c.cslq.Encode("c c", localListenOffers, filter)
	if err != nil {
		err = Error(err, "Cannot request listen offers")
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
		err = Error(err, "Cannot update peer")
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
