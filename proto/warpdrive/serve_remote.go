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
		case <-d.ctx.Done():
		case <-finish:
		}
		_ = d.conn.Close()
	}()
	for {
		var code byte
		if err = cslq.Decode(d.conn, "c", &code); err != nil {
			err = Error(err, "Cannot read ping")
			return
		}
		if err = d.cslq.Encode("c", code); err != nil {
			err = Error(err, "Cannot write ping")
			return
		}
	}
}

func (d Dispatcher) Receive() (err error) {
	peerId := PeerId(d.callerId)
	peer := d.srv.Peer().Get(peerId)
	// Check if peer is blocked
	if peer.Mod == PeerModBlock {
		d.conn.Close()
		d.log.Println("Blocked request from", peerId)
		return
	}
	// Read file offer id
	var offerId OfferId
	err = d.cslq.Decode("[c]c", &offerId)
	if err != nil {
		err = Error(err, "Cannot read offer id")
		return
	}
	// Read files request
	var files []Info
	err = json.NewDecoder(d.conn).Decode(&files)
	if err != nil {
		err = Error(err, "Cannot read files for offer", offerId)
		return
	}
	// Store incoming offer
	d.srv.Incoming().Add(offerId, files, peerId)
	// Auto accept offer if peer is trusted
	code := OfferAwaiting
	if peer.Mod == PeerModTrust {
		err = d.Download(offerId)
		if err != nil {
			d.log.Println("Cannot auto accept files offer", offerId, err)
		} else {
			code = OfferAccepted
		}
	}
	// Send received
	_ = d.cslq.Encode("c", code)
	return
}

func (d Dispatcher) Upload(
	offerId OfferId,
	index int,
	offset int64,
) (err error) {
	srv := d.srv.Outgoing()

	// Obtain setup service with offer id
	offer := srv.Get(offerId)
	if offer == nil {
		err = Error(nil, "Cannot find offer with id", offerId)
		return
	}

	// Update status
	srv.Accept(offer)

	// Send confirmation
	err = d.cslq.Encode("c", 0)
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
		err := d.srv.File().Copy(offer).To(d.conn)
		if err != nil {
			err = Error(err, "Cannot upload files")
		}
		finish <- err
	}()

	// Read OK
	go func() {
		var code byte
		if err := d.cslq.Decode("c", &code); err != nil {
			err = Error(err, "Cannot read ok")
			d.log.Println(err)
			_ = d.conn.Close()
		}
	}()

	// Ensure the status will be updated
	func() {
		d.job.Add(1)
		select {
		case err = <-finish:
		case <-d.ctx.Done():
			_ = d.conn.Close()
			err = <-finish
		}
		srv.Finish(offer, err)
		time.Sleep(200)
		d.job.Done()
	}()
	return
}
