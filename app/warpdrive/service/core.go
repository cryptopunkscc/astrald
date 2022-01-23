package warpdrive

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/ioprogress"
	"io"
	"log"
	"time"
)

type core struct {
	Config
	*log.Logger
	*persistence
	*cache
	*observers
}

type cache struct {
	offers
	peers Peers
}

type offers struct {
	incoming Offers
	outgoing Offers
}

type persistence struct {
	repo     repository
	resolver Resolver
	received storage
}

type observers struct {
	filesOffers    *subscriptions
	incomingStatus *subscriptions
	outgoingStatus *subscriptions
}

// ================================ Setup =================================

func (c *core) setupCore() {
	c.persistence = &persistence{}
	c.cache = &cache{}
	c.observers = &observers{}
	c.filesOffers = newSubscriptions()
	c.incomingStatus = newSubscriptions()
	c.outgoingStatus = newSubscriptions()
}

func (c *core) setupResolver() {
	if c.RemoteResolver {
		c.resolver = newRemoteResolver()
	} else {
		c.resolver = newDefaultResolver()
	}
}

func (c *core) setupStorage() {
	c.received = receivedFilesStorage()
}

func (c *core) setupRepository() {
	if c.RepositoryDir != "" {
		c.repo = newRepository(newStorage(c.RepositoryDir))
	} else {
		c.repo = newRepository(filesStorage())
	}
}

func (c *core) setPeers(peers Peers) {
	c.peers = peers
}

func (c *core) setupPeers() {
	peers := c.repo.listPeers()
	for _, peer := range peers {
		peerRef := peer
		c.peers[peer.Id] = &peerRef
	}
}

func (c *core) setupOffers() {
	c.incoming = c.repo.listIncoming()
	c.outgoing = c.repo.listOutgoing()
}

// ================================ API =================================

func (c *core) addOutgoingOffer(offerId string, files []Info) {
	// Cache outgoing files request
	offer := &Offer{
		Status: Status{
			Id:     OfferId(offerId),
			Status: "sent",
		},
		Files: files,
	}
	c.outgoing[offer.Id] = offer
	c.repo.saveOutgoing(offer)
	// Notify status listeners
	go c.notifyListeners(offer.Status, c.outgoingStatus)
}

func (c *core) getPeer(id PeerId) Peer {
	peer := c.peers[id]
	if peer == nil {
		peer = &Peer{
			Id:    id,
			Alias: "",
			Mod:   "",
		}
		c.peers[id] = peer
	}
	return *peer
}

func (c *core) addIncomingOffer(peer Peer, offerId string, files []Info) {
	offer := &Offer{
		Files: files,
		Peer:  peer.Id,
		Status: Status{
			Id:     OfferId(offerId),
			Status: "received",
		},
	}

	c.incoming[offer.Id] = offer
	c.repo.saveIncoming(offer)

	go c.notifyListeners(offer.Status, c.incomingStatus)
	if peer.Mod == PEER_MOD_ASK {
		go c.notifyListeners(offer, c.filesOffers)
	}
}

func (c *core) getIncomingOffer(id OfferId) *Offer {
	return c.incoming[id]
}

func (c *core) getOutgoingOffer(id OfferId) *Offer {
	return c.outgoing[id]
}

func (c *core) updateIncomingOfferStatus(offer *Offer, status string, persist bool) {
	offer.Status.Status = status
	if persist {
		c.repo.saveIncoming(offer)
	}
	go c.notifyListeners(offer.Status, c.incomingStatus)
}

func (c *core) updateOutgoingOfferStatus(offer *Offer, status string, persist bool) {
	offer.Status.Status = status
	if persist {
		c.repo.saveOutgoing(offer)
	}
	go c.notifyListeners(offer.Status, c.outgoingStatus)
}

func (c *core) updatePeer(peerId string, attr string, value string) {
	peer := c.peers[PeerId(peerId)]
	if peer == nil {
		c.Println("Cannot find peer with id", peerId)
		return
	}
	switch attr {
	case "mod":
		peer.Mod = value
	case "alias":
		peer.Alias = value
	default:
		c.Println("Invalid peer attribute", attr)
		return
	}
	var peers []Peer
	for _, p := range c.peers {
		peers = append(peers, *p)
	}
	c.repo.savePeers(peers)
}

func (c *core) listPeers() (peers []*Peer) {
	peers = make([]*Peer, len(c.peers))
	i := 0
	peersMap := c.peers
	for key := range peersMap {
		peers[i] = c.peers[key]
		i++
	}
	return
}

func (c *core) copyFilesFrom(reader io.Reader, offer *Offer) (err error) {
	for _, file := range offer.Files {
		if err = c.copyFileFrom(reader, offer, file); err != nil {
			return
		}
	}
	return
}

func (c *core) copyFileFrom(reader io.Reader, offer *Offer, file Info) (err error) {
	// Obtain writer
	if file.IsDir {
		err := c.received.MkDir(file.Path, file.Perm)
		if err != nil && !c.received.IsExist(err) {
			c.Println("Cannot make dir", file.Path, err)
			return err
		}
	} else {
		writer, err := c.received.Writer(file.Path, file.Perm)
		if err != nil {
			c.Println("Cannot get writer for", file.Path, err)
			return err
		}
		defer writer.Close()
		// Copy bytes
		progress := &ioprogress.Reader{
			Reader:       reader,
			Size:         file.Size,
			DrawInterval: 200 * time.Millisecond,
			DrawFunc: func(progress int64, size int64) error {
				status := fmt.Sprintf("download: %s %d/%dB", file.Path, progress, size)
				c.updateIncomingOfferStatus(offer, status, false)
				return nil
			},
		}
		_, err = io.CopyN(writer, progress, file.Size)
		if err != nil {
			c.Println("Cannot copy", file.Path, err)
			return err
		}
		err = writer.Close()
		if err != nil {
			c.Println("Cannot close file", file.Path, err)
			return err
		}
	}
	return err
}

func (c *core) copyFilesTo(writer io.Writer, offer *Offer) (err error) {
	for _, file := range offer.Files {
		if file.IsDir {
			continue
		}
		if err = c.copyFileTo(writer, offer, file); err != nil {
			return
		}
	}
	return
}

func (c *core) copyFileTo(writer io.Writer, offer *Offer, file Info) (err error) {
	reader, err := c.resolver.Reader(file.Path)
	if err != nil {
		c.Println("Cannot get reader", file.Path, offer.Id, err)
		return
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         file.Size,
		DrawInterval: 200 * time.Millisecond,
		DrawFunc: func(progress int64, size int64) error {
			status := fmt.Sprintf("upload %s %d/%dB", file.Path, progress, size)
			c.updateOutgoingOfferStatus(offer, status, false)
			return nil
		},
	}
	_, err = io.CopyN(writer, progress, file.Size)
	if err != nil {
		c.Println("Cannot copy", file.Path, err)
		return
	}
	return
}

// ================================ Utils ================================

func (c *core) notifyListeners(data interface{}, listeners *subscriptions) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.Println("Cannot create json from data", data, err)
		return
	}
	for listener := range listeners.set {
		_, err := listener.Write(jsonData)
		if err != nil {
			c.Println("Error while sending files to listener", err)
			return
		}
	}
}
