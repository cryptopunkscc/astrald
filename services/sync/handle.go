package sync

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

func (srv service) Download(rc request.Context) error {

	// Get node id
	log.Println(rc.Port, "getting node id for sync")
	nodeId, err := rc.ReadStringWithSize8()
	if err != nil {
		log.Println(rc.Port, "cannot get node id for sync", err)
		return err
	}

	// Get file id
	log.Println(rc.Port, "getting file id for sync")
	fileId, _, err := fid.Read(rc)
	if err != nil {
		log.Println(rc.Port, "cannot get file id for sync", err)
		return err
	}

	// Download file
	log.Println(rc.Port, "downloading file", nodeId, fileId, err)
	remoteFiles := repo.NewFilesClient(srv, srv, api.Identity(nodeId))
	err = download(srv, remoteFiles, fileId)
	if err != nil {
		log.Println(rc.Port, "cannot download file", nodeId, fileId, err)
		return err
	}

	log.Println(rc.Port, "sending ok", nodeId, fileId, err)
	err = rc.WriteByte(0)
	if err != nil {
		log.Println(rc.Port, "cannot sending ok", nodeId, fileId, err)
		return err
	}
	log.Println(rc.Port, "finish download file", nodeId, fileId, err)
	return nil
}
