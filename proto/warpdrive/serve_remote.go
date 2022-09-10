package warpdrive

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/cslq"
	"time"
)

func (d Dispatcher) Ping() (err error) {
	finish := make(chan struct{})
	defer close(finish)
	go func() {
		select {
		case <-d.Done():
		case <-finish:
		}
		_ = d.Conn.Close()
	}()
	for {
		var code byte
		if err = cslq.Decode(d.Conn, "c", &code); err != nil {
			err = Error(err, "Cannot read ping")
			return
		}
		if err = d.Encode("c", code); err != nil {
			err = Error(err, "Cannot write ping")
			return
		}
	}
}

func (d Dispatcher) Receive() (err error) {
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
	var files []Info
	err = json.NewDecoder(d.Conn).Decode(&files)
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

func (d Dispatcher) Upload(
	offerId OfferId,
	index int,
	offset int64,
) (err error) {
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

	finish := make(chan error)

	// Send files
	go func() {
		defer close(finish)
		// Copy files to connection
		offer.Index = index
		offer.Progress = offset
		err := d.File().Copy(offer).To(d.Conn)
		if err != nil {
			err = Error(err, "Cannot upload files")
		}
		finish <- err
	}()

	// Read OK
	go func() {
		var code byte
		if err := d.Decode("c", &code); err != nil {
			err = Error(err, "Cannot read ok")
			d.Println(err)
			_ = d.Conn.Close()
		}
	}()

	// Ensure the status will be updated
	func() {
		d.Job.Add(1)
		select {
		case err = <-finish:
		case <-d.Done():
			_ = d.Conn.Close()
			err = <-finish
		}
		srv.Finish(offer, err)
		time.Sleep(200)
		d.Job.Done()
	}()
	return
}
