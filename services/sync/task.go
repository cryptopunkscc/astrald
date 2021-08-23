package sync

import (
	"github.com/cryptopunkscc/astrald/services/identity"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/share"
	"log"
	"time"
)

func (rc *requestContext) syncLoop() {
	for rc.Err() == nil {
		time.Sleep(10 * time.Second)
		rc.sync()
	}
}

func (rc *requestContext) sync() {

	// Getting contacts for sync
	log.Println(Port, "listing contacts")
	cards, err := identity.NewClient(rc, rc).List()
	if err != nil {
		log.Println(Port, "cannot list contacts", err)
		return
	}

	// Handle each contact
	for _, card := range cards {
		remoteId := card.Id
		go func() {
			// get list of shared files
			files, err := share.NewSharesClient(rc, rc, remoteId).List()
			if err != nil {
				return
			}
			// handle each shared file id
			filesClient := repo.NewFilesClient(rc, rc, remoteId)
			for _, fileId := range files {
				err := download(rc, filesClient, fileId)
				if err != nil {
					log.Println(Port, "cannot sync", fileId, err)
				}
			}
		}()
	}
}
