package handle

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/services/push/internal/service"
	"github.com/cryptopunkscc/astrald/services/sync"
	"log"
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
		err := sync.NewClient(r.Ctx, r.Core).Download(r.Caller, id)
		if err != nil {
			log.Println(r.Port, "cannot download file with id", id.String())
		}
		log.Println(r.Port, "notifying observers about id", id.String())
		notifyObservers(r, idBuff)
	}()
	return nil
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
