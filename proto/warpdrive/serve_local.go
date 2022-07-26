package warpdrive

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"time"
)

func (d Dispatcher) Ping() (err error) {
	for {
		var code byte
		if err = cslq.Decode(d.Conn, "c", &code); err != nil {
			return Error(err, "Cannot read ping")
		}
		if err = d.Encode("c", code); err != nil {
			return Error(err, "Cannot write ping")
		}
	}
}

func (d *Dispatcher) send(peerId PeerId, filePath string) (err error) {
	// Get files info
	files, err := d.File().Info(filePath)
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
	client, err := NewClient(d.Api).ConnectRemote(identity)
	if err != nil {
		err = Error(err, "Cannot connect to remote", peerId)
		return
	}

	// Send file to recipient service
	offerId, code, err := client.send(files)
	_ = client.Close()
	if err != nil {
		err = Error(err, "Cannot send file")
		return
	}

	d.Outgoing().Add(offerId, files, peerId)

	// Write id to sender
	err = d.Encode("[c]c", offerId)
	if err != nil {
		err = Error(err, "Cannot send id", offerId)
		return
	}
	// Write code to sender
	err = d.Encode("c", code)
	if err != nil {
		err = Error(err, "Cannot code", offerId)
		return
	}
	d.Println(filePath, "offer sent to", peerId)
	return
}

func (d *Dispatcher) offers(filter Filter) (err error) {
	// Collect file offers
	offers := d.filterOffers(filter)
	// Send filtered file offers
	if err = json.NewEncoder(d.Conn).Encode(offers); err != nil {
		err = Error(err, "Cannot send incoming offers")
		return
	}
	// Wait for OK
	var code byte
	if err = cslq.Decode(d.Conn, "c", &code); err != nil {
		err = Error(err, "Cannot read ok")
		return
	}
	d.Println("Sent", filter, "offers")
	return
}

func (d Dispatcher) Download(offerId OfferId) (err error) {
	// Get incoming offer service for offer id
	srv := d.Incoming()
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
	client, err := NewClient(d.Api).ConnectRemote(peerId)
	if err != nil {
		return
	}

	// Request download
	if err = client.Download(offerId); err != nil {
		err = Error(err, "Cannot download offer")
		return err
	}

	// download files in background
	go func(c RemoteClient) {
		// Ensure the status will be updated
		var err error
		defer func() {
			_ = c.Close()
			time.Sleep(200 * time.Millisecond)
			srv.Finish(offer, err)
		}()

		// Copy files to storage
		err = d.File().Copy(offer).From(c.Conn)
		if err != nil {
			err = Error(err, "Cannot download files")
			return
		}

		// Send OK
		err = c.Encode("c", 0)
		if err != nil {
			err = Error(err, "Cannot send ok")
			return
		}
	}(client)
	return
}

func (d *Dispatcher) Accept(offerId OfferId) (err error) {
	// Download offer
	d.Println("Accepted incoming files", offerId)
	err = d.Download(offerId)
	if err != nil {
		err = Error(err, "Cannot download incoming files", offerId)
		return
	}
	// Send ok
	err = d.Encode("c", 0)
	if err != nil {
		err = Error(err, "Cannot send ok")
		return
	}
	return
}

func (d *Dispatcher) peers() (err error) {
	// Get peers
	peers := d.Peer().List()
	// Send peers
	if err = json.NewEncoder(d.Conn).Encode(peers); err != nil {
		err = Error(err, "Cannot send peers")
		return
	}
	// Read OK
	var code byte
	if err = cslq.Decode(d.Conn, "c", &code); err != nil {
		err = Error(err, "Cannot read ok")
		return
	}
	return
}

func (d *Dispatcher) status(filter Filter) (err error) {
	unsub := d.filterSubscribe(filter, OfferService.StatusSubscriptions)
	defer unsub()
	// Wait for close
	var code byte
	err = cslq.Decode(d.Conn, "c", &code)
	return
}

func (d *Dispatcher) subscribe(filter Filter) (err error) {
	unsub := d.filterSubscribe(filter, OfferService.OfferSubscriptions)
	defer unsub()
	// Wait for close
	var code byte
	err = cslq.Decode(d.Conn, "c", &code)
	return
}

func (d *Dispatcher) update() (err error) {
	// Read peer update
	// Fixme refactor to cslq
	var req []string
	if err = json.NewDecoder(d.Conn).Decode(&req); err != nil {
		err = Error(err, "Cannot read peer update")
		return
	}
	peerId := req[0]
	attr := req[1]
	value := req[2]
	// Update peer
	d.Peer().Update(peerId, attr, value)
	// Send OK
	if err = d.Encode("c", 0); err != nil {
		err = Error(err, "Cannot send ok")
		return
	}
	return
}
