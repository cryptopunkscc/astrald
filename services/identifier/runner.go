package identifier

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/serialize"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/identifier/internal"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/register"
	"log"
	"time"
)

func init() {
	_ = node.RegisterService(Port, run)
}

const Port = "identifier"

var resolvers = []Resolve{
	internal.GetStoryType,
	internal.GetMimeType,
}

type Resolve func(prefix []byte) (string, error)

func run(ctx context.Context, core api.Core) error {
	repository := repo.NewRepoClient(ctx, core)
	observers := map[api.Stream]string{}

	// Observe repo changes
	go func() {
		time.Sleep(1 * time.Second)

		// Request observe
		stream, err := repository.Observer()
		if err != nil {
			log.Println(Port, "cannot connect to ", repo.Port, err)
			return
		}

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
				reader, err := repository.Reader(id)
				defer reader.Close()
				if err != nil {
					log.Println(Port, "cannot read", err)
					return
				}

				// obtain file prefix
				log.Println(Port, "reading", id.Size, "bytes from", repo.Port)
				prefixBuff, err := serialize.NewParser(reader).ReadN(4096)
				if err != nil {
					log.Println(Port, "cannot read from", repo.Port, err)
					return
				}
				log.Println(Port, "resolved file prefix")

				// resolve file type
				var fileType string
				for _, resolve := range resolvers {
					fileType, err = resolve(prefixBuff[:])
					if err == nil {
						break
					}
				}
				if err != nil || fileType == "" {
					log.Println(Port, "cannot resolve fileType")
					return
				}
				log.Println(Port, "resolved file type", fileType)

				// notify observers
				log.Println(Port, "notifying observers", len(observers))
				for observer, observedType := range observers {
					if observedType == fileType {
						go func() {
							_, err := observer.Write(idBuff[:])
							if err != nil {
								log.Println(Port, "cannot write file id for", observedType, err)
								return
							}
						}()
					}
				}
			}()
		}
	}()

	// Handle incoming connections
	handler, err := register.Port(ctx, core, Port)
	if err != nil {
		return err
	}
	for request := range handler.Requests() {
		stream := accept.Request(ctx, request)
		log.Println(Port, "accepted new connection")

		go func() {
			defer stream.Close()

			// Read query
			size, err := stream.ReadByte()
			if err != nil {
				return
			}

			query, err := stream.ReadString(int(size))
			if err != nil {
				return
			}

			// Register observer
			observers[stream] = query
			log.Println(Port, "added new files observer for", query)

			// Close blocking
			for {
				_, err := stream.ReadByte()
				if err != nil {
					log.Println(Port, "removing file observer")
					delete(observers, stream)
					return
				}
			}
		}()
	}
	return nil
}
