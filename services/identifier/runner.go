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
	"io"
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
		var err error
		var stream io.Reader
		var id fid.ID
		var idBuff [fid.Size]byte

		// Request observe
		if stream, err = repository.Observer(); err != nil {
			log.Println(Port, "cannot connect to ", repo.Port, err)
			return
		}

		for {
			// Read id
			if id, idBuff, err = fid.Read(stream); err != nil {
				log.Println(Port, "cannot read new fid from repo", err)
				return
			}
			log.Println(Port, "new file fid", id.String())
			// handle received id
			go func() {
				var err error
				var reader io.ReadCloser
				var prefixBuff []byte
				var fileType string

				// obtain file reader for id
				if reader, err = repository.Reader(id); err != nil {
					log.Println(Port, "cannot read", err)
					return
				} else {
					defer reader.Close()
				}

				// obtain file prefix
				log.Println(Port, "reading", id.Size, "bytes from", repo.Port)
				if prefixBuff, err = serialize.NewParser(reader).ReadN(4096); err != nil {
					log.Println(Port, "cannot read from", repo.Port, err)
					return
				} else {
					log.Println(Port, "resolved file prefix")
				}

				// resolve file type
				for _, resolve := range resolvers {
					fileType, err = resolve(prefixBuff[:])
					if err == nil {
						break
					}
				}
				if err != nil || fileType == "" {
					log.Println(Port, "cannot resolve fileType")
					return
				} else {
					log.Println(Port, "resolved file type", fileType)
				}

				// notify observers
				log.Println(Port, "notifying observers", len(observers))
				for observer, observedType := range observers {
					if observedType == fileType {
						go func() {
							if _, err := observer.Write(idBuff[:]); err != nil {
								log.Println(Port, "cannot write file id for", observedType, err)
							}
						}()
					}
				}
			}()
		}
	}()

	// Handle incoming connections
	var err error
	var handler api.PortHandler
	if handler, err = register.Port(ctx, core, Port); err != nil {
		return err
	}
	for request := range handler.Requests() {
		stream := accept.Request(ctx, request)
		log.Println(Port, "accepted new connection")

		// Handler connection
		go func() {
			if err := func() (err error) {
				defer stream.Close()

				var size byte
				var query string

				// Read query
				if size, err = stream.ReadByte(); err != nil {
					return
				}
				if query, err = stream.ReadString(int(size)); err != nil {
					return
				}

				// Register observer
				observers[stream] = query
				log.Println(Port, "added new files observer for", query)

				// Close blocking
				for {
					if _, err = stream.ReadByte(); err != nil {
						log.Println(Port, "removing file observer")
						delete(observers, stream)
						return
					}
				}
			}(); err != nil {
				log.Println(Port, "cannot handle connection", err)
			}
		}()
	}
	return nil
}
