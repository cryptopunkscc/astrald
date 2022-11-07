package warpdrive

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"time"
)

func (d Dispatcher) CreateOffer(peerId PeerId, filePath string) (err error) {
	// Get files info
	files, err := d.srv.File().Info(filePath)
	if err != nil {
		err = Error(err, "Cannot get files info")
		return
	}

	// Parse identity
	identity, err := id.ParsePublicKeyHex(string(peerId))
	if err != nil {
		err = Error(err, "Cannot parse peer id")
		return
	}

	// Connect to remote client
	client, err := NewClient(d.api).Connect(identity, Port)
	if err != nil {
		err = Error(err, "Cannot connect to remote", peerId)
		return
	}

	// Send file to recipient service
	offerId, code, err := client.SendOffer(files)
	_ = client.Close()
	if err != nil {
		err = Error(err, "Cannot send file")
		return
	}

	d.srv.Outgoing().Add(offerId, files, peerId)

	// Write id to sender
	err = d.cslq.Encode("[c]c c", offerId, code)
	if err != nil {
		err = Error(err, "Cannot send create offer result", offerId)
		return
	}
	d.log.Println(filePath, "offer sent to", peerId)
	return
}

func (d Dispatcher) ListOffers(filter Filter) (err error) {
	// Collect file offers
	offers := d.filterOffers(filter)
	d.log.Println("Filter", filter)
	// Send filtered file offers
	if err = json.NewEncoder(d.conn).Encode(offers); err != nil {
		err = Error(err, "Cannot send incoming offers")
		return
	}
	return
}

func (d Dispatcher) AcceptOffer(offerId OfferId) (err error) {
	// Download offer
	d.log.Println("Accepted incoming files", offerId)
	err = d.Download(offerId)
	if err != nil {
		err = Error(err, "Cannot download incoming files", offerId)
		return
	}
	// Send ok
	err = d.cslq.Encode("c", 0)
	if err != nil {
		err = Error(err, "Cannot send ok")
		return
	}
	return
}

func (d Dispatcher) Download(offerId OfferId) (err error) {
	// Get incoming offer service for offer id
	srv := d.srv.Incoming()
	offer := srv.Get(offerId)
	if offer == nil {
		err = Error(nil, "Cannot find incoming file")
		return
	}

	// parse peer id
	peerId, err := id.ParsePublicKeyHex(string(offer.Peer))
	if err != nil {
		err = Error(err, "Cannot parse peer id")
		return
	}

	// Update status
	srv.Accept(offer)

	// Connect to remote warpdrive
	client, err := NewClient(d.api).Connect(peerId, Port)
	if err != nil {
		return
	}

	// Request download
	if err = client.Download(offerId, offer.Index, offer.Progress); err != nil {
		err = Error(err, "Cannot download offer")
		return err
	}

	finish := make(chan error)

	// Ensure the status will be updated
	go func() {
		d.job.Add(1)
		select {
		case err = <-finish:
		case <-d.ctx.Done():
			_ = client.conn.Close()
			err = <-finish
		}
		if err == nil {
			_ = client.Close()
		}
		srv.Finish(offer, err)
		d.log.Println(Error(err, "Failed"))
		time.Sleep(200)
		d.job.Done()
	}()

	// download files in background
	go func() {
		defer close(finish)
		// Copy files from connection to storage
		if err = srv.Copy(offer).From(client.conn); err != nil {
			finish <- Error(err, "Cannot download files")
			return
		}
		// Send OK
		if err = client.cslq.Encode("c", 0); err != nil {
			finish <- Error(err, "Cannot send ok")
			return
		}
		finish <- nil
	}()
	return
}

func (d Dispatcher) ListPeers() (err error) {
	// Get peers
	peers := d.srv.Peer().List()
	// Send peers
	if err = json.NewEncoder(d.conn).Encode(peers); err != nil {
		err = Error(err, "Cannot send peers")
		return
	}
	return
}

func (d Dispatcher) ListenStatus(filter Filter) (err error) {
	unsub := d.filterSubscribe(filter, OfferService.StatusSubscriptions)
	defer unsub()
	// Wait for close
	var code byte
	err = d.cslq.Decode("c", &code)
	return
}

func (d Dispatcher) ListenOffers(filter Filter) (err error) {
	unsub := d.filterSubscribe(filter, OfferService.OfferSubscriptions)
	defer unsub()
	// Wait for close
	var code byte
	err = d.cslq.Decode("c", &code)
	return
}

func (d Dispatcher) UpdatePeer() (err error) {
	// Read peer update
	// Fixme refactor to cslq
	var req []string
	if err = json.NewDecoder(d.conn).Decode(&req); err != nil {
		err = Error(err, "Cannot read peer update")
		return
	}
	peerId := req[0]
	attr := req[1]
	value := req[2]
	// Update peer
	d.srv.Peer().Update(peerId, attr, value)
	// Send OK
	if err = d.cslq.Encode("c", 0); err != nil {
		err = Error(err, "Cannot send ok")
		return
	}
	return
}
