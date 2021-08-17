package identifier

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/serialize"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/identifier/internal"
	"github.com/cryptopunkscc/astrald/services/repo"
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
	observers := map[api.Stream]string{}
	network := core.Network()

	// Observe repo changes
	go func() {
		time.Sleep(1 * time.Second)

		// Connect
		conn, err := network.Connect("", repo.Port)
		if err != nil {
			log.Println(Port, "cannot connect to ", repo.Port, err)
			return
		}
		s := serialize.NewSerializer(conn)
		go func() {
			<-ctx.Done()
			_ = conn.Close()
		}()

		// Request observe
		err = s.WriteByte(repo.RequestObserve)
		if err != nil {
			log.Println(Port, "cannot observe", repo.Port, err)
			return
		}

		// Read ids
		for {
			var idBuff [fid.Size]byte
			_, err := s.Read(idBuff[:])
			if err != nil {
				log.Println(Port, "cannot read new fid from repo", err)
				return
			}
			id := fid.Unpack(idBuff)

			// handle received id
			go func() {

				// obtain file prefix
				log.Println(Port, "new file fid", id.String())
				stream, err := network.Connect("", repo.Port)
				if err != nil {
					log.Println(Port, "cannot connect to", repo.Port, err)
					return
				}
				defer stream.Close()

				// Request file read
				s := serialize.NewSerializer(stream)
				err = s.WriteByte(repo.RequestRead)
				if err != nil {
					log.Println(Port, "cannot read", err)
					return
				}

				// Write file id
				_, err = s.Write(idBuff[:])
				if err != nil {
					log.Println(Port, "cannot write file id", err)
					return
				}

				log.Println(Port, "reading", id.Size, "bytes from", repo.Port)
				prefixBuff, err := s.ReadN(4096)
				if err != nil {
					log.Println(Port, "cannot read from", repo.Port, err)
					return
				}
				log.Println(Port, "resolved file prefix")

				// resolve file type
				fileType := ""
				for _, resolve := range resolvers {
					fileType, err = resolve(prefixBuff[:])
					if err == nil {
						break
					}
				}
				if fileType == "" {
					log.Println(Port, "cannot resolve fileType")
					return
				}
				log.Println(Port, "resolved file type", fileType)

				// notify observers
				log.Println(Port, "notifying observers", len(observers))
				for stream, observedType := range observers {
					if observedType == fileType {
						go func() {
							_, err := stream.Write(idBuff[:])
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

	handler, err := core.Network().Register(Port)
	go func() {
		<-ctx.Done()
		_ = handler.Close()
	}()
	if err != nil {
		return err
	}

	// Handle incoming connections
	for conn := range handler.Requests() {
		stream := conn.Accept()
		log.Println(Port, "accepted new connection")

		go func() {
			defer stream.Close()
			s := serialize.NewSerializer(stream)

			// Read query
			query, err := s.ReadStringWithSize()
			if err != nil {
				return
			}

			// Register observer
			observers[stream] = query
			log.Println(Port, "added new files observer for", query)

			// Close blocking
			var buffer [1]byte
			for {
				_, err := stream.Read(buffer[:])
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
