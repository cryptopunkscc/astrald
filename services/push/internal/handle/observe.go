package handle

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	repo2 "github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/services/push/internal/service"
	"github.com/cryptopunkscc/astrald/services/repo"
	"io"
	"log"
)

func Push(r *service.Request) (err error) {
	for {
		var idBuff [fid.Size]byte

		// Read next id
		if _, err = r.Read(idBuff[:]); err != nil {
			log.Println(r.Port, "cannot read file id", err)
			return
		}

		// Handle id
		id := fid.Unpack(idBuff)
		go func() { _ = downloadFile(r, id) }()
	}
}

func downloadFile(r *service.Request, id fid.ID) (err error) {
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

func Observe(r *service.Request) (err error) {
	r.Observers[r] = struct{}{}
	for {
		if _, err = r.ReadByte(); err != nil {
			log.Println(r.Port, "removing file observer")
			delete(r.Observers, r)
			return
		}
	}
}
