package lore

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/serialize"
	lore "github.com/cryptopunkscc/astrald/components/story"
	"github.com/cryptopunkscc/astrald/node"
	_fs "github.com/cryptopunkscc/astrald/services/fs"
	"github.com/cryptopunkscc/astrald/services/identifier"
	"log"
	"time"
)

const Port = "lore"

func init() {
	_ = node.RegisterService(Port, run)
}

const storyMimeType = "application/lore"

func run(ctx context.Context, core api.Core) error {
	observers := map[api.Stream]string{}
	network := core.Network()

	go func() {
		time.Sleep(1 * time.Second)

		// Connect to identifier
		conn, err := network.Connect("", identifier.Port)
		if err != nil {
			log.Println(Port, "cannot connect", identifier.Port, err)
			return
		}
		log.Println(Port, "connected to", identifier.Port, err)
		go func() {
			<-ctx.Done()
			_ = conn.Close()
		}()
		s := serialize.NewSerializer(conn)

		// Send observed type
		_, err = s.WriteStringWithSize(storyMimeType)
		if err != nil {
			log.Println(Port, "cannot request observe", identifier.Port, err)
			return
		}
		log.Println(Port, "observing", identifier.Port, err)

		// Handle incoming ids
		for {
			var idBuff [40]byte

			// Resolving id
			_, err := conn.Read(idBuff[:])
			if err != nil {
				log.Println(Port, "read new fid from", _fs.Port, err)
				return
			}
			id := fid.Unpack(idBuff)
			log.Println(Port, "new file fid", id.String())

			go func() {
				// Connecting to fs
				fs, err := network.Connect("", _fs.Port)
				if err != nil {
					log.Println(Port, "cannot connect to", _fs.Port, err)
					return
				}

				s := serialize.NewSerializer(fs)

				// Request read
				err = s.WriteByte(_fs.RequestRead)
				if err != nil {
					log.Println(Port, "cannot read from", _fs.Port, err)
					return
				}

				// Sending file id
				_, err = s.Write(idBuff[:])
				if err != nil {
					log.Println(Port, "cannot write file id", _fs.Port, err)
					return
				}

				// Read story
				story, err := lore.Unpack(s)
				if err != nil {
					log.Println(Port, "cannot unpack story", err)
					return
				}
				storyType := story.Type()
				log.Println(Port, "resolved story type", storyType, "from", _fs.Port)

				// Notify observers
				for observer, observerTyp := range observers {
					if storyType == observerTyp {
						err = story.Pack(observer)
						if err != nil {
							log.Println(Port, "cannot send story for", storyType, err)
							return
						}
						log.Println(Port, "send story for", storyType)
					}
				}
			}()
		}
	}()

	// Handle incoming connections
	handler, err := core.Network().Register(Port)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = handler.Close()
	}()
	for conn := range handler.Requests() {
		stream := conn.Accept()
		log.Println(Port, "accepted new connection")

		go func() {
			defer stream.Close()
			s := serialize.NewSerializer(stream)

			// Read type
			typ, err := s.ReadStringWithSize()
			if err != nil {
				return
			}

			// Register observer
			observers[stream] = typ
			log.Println(Port, "added new files observer for", typ)

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
