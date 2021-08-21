package handle

import (
	"errors"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	repo2 "github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/services/push/internal/service"
	"github.com/cryptopunkscc/astrald/services/repo"
	"io"
	"log"
	"time"
)

func Push(r *service.Request) error {
	ok := []byte{0}
	log.Println(r.Port, "reading file id")
	id, idBuff, err := fid.Read(r)
	if err != nil {
		log.Println(r.Port, "cannot read file id", err)
		return err
	}
	log.Println(r.Port, "sending ok")
	_, err = r.Write(ok)
	go func() {
		log.Println(r.Port, "downloading file with id", id.String())
		_ = downloadFile(r, id)
		log.Println(r.Port, "notifying observers about id", id.String())
		notifyObservers(r, idBuff)
	}()
	return nil
}

func downloadFile(
	r *service.Request,
	id fid.ID,
) (err error) {
	var reader repo2.Reader
	var writer repo2.Writer

	// Obtain remote reader
	log.Println(r.Port, "getting files reader for id", id.String())
	if reader, err = repo.NewFilesClient(r.Ctx, r, r.Caller).Reader(id); err != nil {
		log.Println(r.Port, "cannot obtain remote reader", err)
		return
	}

	// Obtain local writer
	log.Println(r.Port, "getting repo writer", id.String())
	if writer, err = repo.NewRepoClient(r.Ctx, r).Writer(); err != nil {
		log.Println(r.Port, "cannot obtain local writer", err)
		return
	}

	// Write file size
	log.Println(r.Port, "writing file size", id.String())
	_, err = writer.WriteUInt32(uint32(id.Size))
	if err != nil {
		log.Println(r.Port, "cannot write file size", id.Size, id.String())
		return err
	}

	time.Sleep(1000 * time.Millisecond)

	// Copy file into local file system
	defer func() { _ = reader.Close() }()
	defer func() { _, _ = writer.Finalize() }()
	log.Println(r.Port, "coping file to local repo file with size", id.Size, id.String())
	if l, err := io.CopyN(writer, reader, int64(id.Size)); err != nil {
		log.Println(r.Port, "cannot copy file", l, err)
		return err
	}

	// Obtain calculated id for a saved file
	log.Println(r.Port, "getting file id")
	id2, err := writer.Finalize()
	if err != nil {
		log.Println(r.Port, "cannot finalize file copy", err)
		return err
	}

	// Verify calculated id against the received
	log.Println(r.Port, "verifying ids")
	rid := id.String()
	cid := id2.String()
	if rid != cid {
		return errors.New("received id " + rid + " is different than calculated " + cid )
	}

	// Finish
	log.Println(r.Port, "finish coping file to local repo", id.String())
	return
}

func notifyObservers(
	r *service.Request,
	idBuff [fid.Size]byte,
) {
	for observer := range r.Observers {
		go func(s api.Stream) {
			if _, err := s.Write(idBuff[:]); err != nil {
				log.Println(r.Port, "push file id", err)
			}
		}(observer)
	}
}
