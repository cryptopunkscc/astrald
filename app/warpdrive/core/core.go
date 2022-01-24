package core

import (
	"encoding/json"
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/mitchellh/ioprogress"
	"io"
	"log"
	"time"
)

func New(config service.Config) api.Core {
	c := &core{Config: config}
	return c
}

type core struct {
	service.Config
	*log.Logger
	*persistence
	*cache
	*observers
}

type cache struct {
	offers
	peers api.Peers
}

type offers struct {
	incoming api.Offers
	outgoing api.Offers
}

type persistence struct {
	repo repository
	api.Resolver
	received storage
}

type observers struct {
	filesOffers    *api.Subscriptions
	incomingStatus *api.Subscriptions
	outgoingStatus *api.Subscriptions
}

// ================================ Setup =================================

func (c *core) setupCore() {
	c.persistence = &persistence{}
	c.cache = &cache{}
	c.peers = api.Peers{}
	c.observers = &observers{}
	c.filesOffers = api.NewSubscriptions()
	c.incomingStatus = api.NewSubscriptions()
	c.outgoingStatus = api.NewSubscriptions()
}

func (c *core) setupResolver() {
	if c.RemoteResolver {
		c.Resolver = newRemoteResolver()
	} else {
		c.Resolver = newDefaultResolver()
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

func (c *core) setPeers(peers api.Peers) {
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

func (c *core) Setup() {
	c.setupCore()
	c.setupStorage()
	c.setupResolver()
	c.setupRepository()
	c.setupOffers()
	c.setupPeers()
}

func (c *core) SetLogger(logger *log.Logger) {
	c.Logger = logger
}

func (c *core) AddOutgoingOffer(offerId string, files []api.Info) {
	// Cache outgoing files request
	offer := &api.Offer{
		Files: files,
		Status: api.Status{
			Id:     api.OfferId(offerId),
			Status: "sent",
		},
	}
	c.outgoing[offer.Id] = offer
	c.repo.saveOutgoing(offer)
	// Notify status listeners
	go c.notifyListeners(offer.Status, c.outgoingStatus)
}

func (c *core) GetPeer(id api.PeerId) api.Peer {
	peer := c.peers[id]
	if peer == nil {
		peer = &api.Peer{
			Id:    id,
			Alias: "",
			Mod:   "",
		}
		c.peers[id] = peer
	}
	return *peer
}

func (c *core) AddIncomingOffer(peer api.Peer, offerId string, files []api.Info) {
	offer := &api.Offer{
		Files: files,
		Peer:  peer.Id,
		Status: api.Status{
			Id:     api.OfferId(offerId),
			Status: "received",
		},
	}

	c.incoming[offer.Id] = offer
	c.repo.saveIncoming(offer)

	go c.notifyListeners(offer.Status, c.incomingStatus)
	if peer.Mod == api.PeerModAsk {
		go c.notifyListeners(offer, c.filesOffers)
	}
}

func (c *core) GetIncomingOffer(id api.OfferId) *api.Offer {
	return c.incoming[id]
}

func (c *core) GetOutgoingOffer(id api.OfferId) *api.Offer {
	return c.outgoing[id]
}

func (c *core) UpdateIncomingOfferStatus(offer *api.Offer, status string, persist bool) {
	offer.Status.Status = status
	if persist {
		c.repo.saveIncoming(offer)
	}
	go c.notifyListeners(offer.Status, c.incomingStatus)
}

func (c *core) UpdateOutgoingOfferStatus(offer *api.Offer, status string, persist bool) {
	offer.Status.Status = status
	if persist {
		c.repo.saveOutgoing(offer)
	}
	go c.notifyListeners(offer.Status, c.outgoingStatus)
}

func (c *core) UpdatePeer(peerId string, attr string, value string) {
	id := api.PeerId(peerId)
	peer := c.peers[id]
	cached := peer != nil
	if !cached {
		peer = &api.Peer{Id: id}
		c.peers[id] = peer
	}
	switch attr {
	case "mod":
		peer.Mod = value
	case "alias":
		peer.Alias = value
	default:
		if cached {
			return
		}
	}
	var peers []api.Peer
	for _, p := range c.peers {
		peers = append(peers, *p)
	}
	c.repo.savePeers(peers)
}

func (c *core) ListPeers() (peers []*api.Peer) {
	peers = make([]*api.Peer, len(c.peers))
	i := 0
	peersMap := c.peers
	for key := range peersMap {
		peers[i] = c.peers[key]
		i++
	}
	return
}

func (c *core) CopyFilesFrom(reader io.Reader, offer *api.Offer) (err error) {
	for _, file := range offer.Files {
		if err = c.CopyFileFrom(reader, offer, file); err != nil {
			return
		}
	}
	return
}

func (c *core) CopyFileFrom(reader io.Reader, offer *api.Offer, file api.Info) (err error) {
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
				c.UpdateIncomingOfferStatus(offer, status, false)
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

func (c *core) CopyFilesTo(writer io.Writer, offer *api.Offer) (err error) {
	for _, file := range offer.Files {
		if file.IsDir {
			continue
		}
		if err = c.CopyFileTo(writer, offer, file); err != nil {
			return
		}
	}
	return
}

func (c *core) CopyFileTo(writer io.Writer, offer *api.Offer, file api.Info) (err error) {
	reader, err := c.Reader(file.Path)
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
			c.UpdateOutgoingOfferStatus(offer, status, false)
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

func (c *core) GetOutgoingOffers() api.Offers {
	return c.outgoing
}

func (c *core) GetIncomingOffers() api.Offers {
	return c.incoming
}

func (c *core) OutgoingStatus() *api.Subscriptions {
	return c.outgoingStatus
}

func (c *core) IncomingStatus() *api.Subscriptions {
	return c.incomingStatus
}

func (c *core) FilesOffers() *api.Subscriptions {
	return c.filesOffers
}

// ================================ Utils ================================

func (c *core) notifyListeners(data interface{}, listeners *api.Subscriptions) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.Println("Cannot create json from data", data, err)
		return
	}
	listeners.Lock()
	defer listeners.Unlock()
	for listener := range listeners.Set {
		_, err := listener.Write(jsonData)
		if err != nil {
			c.Println("Error while sending files to listener", err)
			return
		}
	}
}
