package warpdrive

import (
	"encoding/json"
	"time"
)

func (d *Dispatcher) Receive() (err error) {
	peerId := PeerId(d.CallerId)
	peer := d.Peer().Get(peerId)
	// Check if peer is blocked
	if peer.Mod == PeerModBlock {
		d.Close()
		d.Println("Blocked request from", peerId)
		return
	}
	// Read file offer id
	var offerId OfferId
	err = d.Decode("[c]c", &offerId)
	if err != nil {
		err = Error(err, "Cannot read offer id")
		return
	}
	// Read files request
	dec := json.NewDecoder(d.Conn)
	var files []Info
	err = dec.Decode(&files)
	if err != nil {
		err = Error(err, "Cannot read files for offer", offerId)
		return
	}
	// Store incoming offer
	d.Incoming().Add(offerId, files, peerId)
	// Auto accept offer if peer is trusted
	code := OfferAwaiting
	if peer.Mod == PeerModTrust {
		err = d.Download(offerId)
		if err != nil {
			d.Println("Cannot auto accept files offer", offerId, err)
		} else {
			code = OfferAccepted
		}
	}
	// Send received
	_ = d.Encode("c", code)
	return
}

func (d *Dispatcher) Upload(offerId OfferId) (err error) {
	srv := d.Outgoing()

	// Obtain setup service with offer id
	offer := srv.Get(offerId)
	if offer == nil {
		err = Error(nil, "Cannot find offer with id", offerId)
		return
	}

	// Update status
	srv.Accept(offer)

	// Send confirmation
	err = d.Encode("c", 0)
	if err != nil {
		err = Error(err, "Cannot send confirmation")
		return
	}

	// Ensure the status will be updated
	defer func() {
		time.Sleep(200 * time.Millisecond)
		srv.Finish(offer, err)
	}()

	// Send files
	err = d.File().Copy(offer).To(d.Conn)
	if err != nil {
		err = Error(err, "Cannot upload files")
		return
	}

	// Read OK
	var code byte
	err = d.Decode("c", &code)
	if err != nil {
		err = Error(err, "Cannot read ok")
		return
	}
	return
}
