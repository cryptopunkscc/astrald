package identifier

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/services/repo"
	"log"
	"time"
)

func (srv service) observeRepo() {
	time.Sleep(1 * time.Second)

	// Request observe
	stream, err := srv.repository.Observer()
	if err != nil {
		log.Println(Port, "cannot connect to", repo.Port, err)
		return
	}

	log.Println(Port, "observing", repo.Port)
	for {
		// Read id
		id, idBuff, err := fid.Read(stream)
		if err != nil {
			log.Println(Port, "cannot read new fid from repo", err)
			return
		}
		log.Println(Port, "new file fid", id.String())

		// handle received id
		go func() {

			// obtain file reader for id
			reader, err := srv.repository.Reader(id)
			if err != nil {
				log.Println(Port, "cannot read", err)
				return
			}
			defer reader.Close()

			// obtain file prefix
			log.Println(Port, "reading", id.Size, "bytes from", repo.Port)
			prefixBuff, err := reader.ReadN(4096)
			if err != nil {
				log.Println(Port, "cannot read from", repo.Port, err)
				return
			}

			// resolve file type
			var fileType string
			log.Println(Port, "resolving file prefix")
			for _, resolve := range srv.resolvers {
				fileType, err = resolve(prefixBuff[:])
				if err == nil {
					break
				}
			}
			if err != nil || fileType == "" {
				log.Println(Port, "cannot resolve fileType")
				return
			}

			// notify observers
			log.Println(Port, "notifying observers", len(srv.observers), "about", fileType)
			for observer, observedType := range srv.observers {
				if observedType == fileType {
					go func(s api.Stream) {
						if _, err := s.Write(idBuff[:]); err != nil {
							log.Println(Port, "cannot write file id for", observedType, err)
						}
					}(observer)
				}
			}
		}()
	}
}
