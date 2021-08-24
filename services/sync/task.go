package sync

import (
	"github.com/cryptopunkscc/astrald/services/identity"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/share"
	"log"
	"time"
)

func (srv *service) syncLoop() {
	for srv.Err() == nil {
		time.Sleep(10 * time.Second)
		srv.sync()
	}
}

func (srv *service) sync() {

	// Getting contacts for sync
	log.Println(Port, "listing contacts")
	cards, err := identity.NewClient(srv, srv).List()
	if err != nil {
		log.Println(Port, "cannot list contacts", err)
		return
	}

	// Handle each contact
	for _, card := range cards {
		remoteId := card.Id
		go func() {
			// get list of shared files
			files, err := share.NewSharesClient(srv, srv, remoteId).List()
			if err != nil {
				return
			}
			// handle each shared file id
			filesClient := repo.NewFilesClient(srv, srv, remoteId)
			for _, fileId := range files {
				err := download(srv, filesClient, fileId)
				if err != nil {
					log.Println(Port, "cannot sync", fileId, err)
				}
			}
		}()
	}
}
