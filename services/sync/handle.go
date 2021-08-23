package sync

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/services/repo"
	"log"
)

func (rc requestContext) Download(
	_ api.Identity,
	query string,
	stream sio.ReadWriteCloser,
) error {

	// Get node id
	log.Println(query, "getting node id for sync")
	nodeId, err := stream.ReadStringWithSize8()
	if err != nil {
		log.Println(query, "cannot get node id for sync", err)
		return err
	}

	// Get file id
	log.Println(query, "getting file id for sync")
	fileId, _, err := fid.Read(stream)
	if err != nil {
		log.Println(query, "cannot get file id for sync", err)
		return err
	}

	// Download file
	log.Println(query, "downloading file", nodeId, fileId, err)
	remoteFiles := repo.NewFilesClient(rc, rc, api.Identity(nodeId))
	err = download(rc, remoteFiles, fileId)
	if err != nil {
		log.Println(query, "cannot download file", nodeId, fileId, err)
		return err
	}

	log.Println(query, "sending ok", nodeId, fileId, err)
	err = stream.WriteByte(0)
	if err != nil {
		log.Println(query, "cannot sending ok", nodeId, fileId, err)
		return err
	}
	log.Println(query, "finish download file", nodeId, fileId, err)
	return nil
}
