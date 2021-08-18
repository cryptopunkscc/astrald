package handle

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	repo2 "github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/services/push/internal/service"
	"github.com/cryptopunkscc/astrald/services/repo"
	"io"
	"log"
)

func Push(r *service.Request) error {
	for {
		id, idBuff, err := fid.Read(r)
		if err != nil {
			log.Println(r.Port, "cannot read file id", err)
			return err
		}
		go func() {
			_ = downloadFile(r, id)
			notifyObservers(r, idBuff)
		}()
	}
}

func downloadFile(
	r *service.Request,
	id fid.ID,
) (err error) {
	var reader repo2.Reader
	var writer repo2.Writer

	// Obtain remote reader
	if reader, err = repo.NewFilesClient(r.Ctx, r, r.Caller).Reader(id); err != nil {
		log.Println(r.Port, "cannot obtain remote reader", err)
		return
	}

	// Obtain local writer
	if writer, err = repo.NewRepoClient(r.Ctx, r).Writer(); err != nil {
		log.Println(r.Port, "cannot obtain local writer", err)
		return
	}

	// Copy file into local file system
	defer func() { _ = reader.Close() }()
	defer func() { _, _ = writer.Finalize() }()
	if _, err = io.Copy(writer, reader); err != nil {
		log.Println(r.Port, "cannot write copy file", err)
		return
	}
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
