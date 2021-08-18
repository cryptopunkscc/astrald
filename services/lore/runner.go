package lore

import (
	"bytes"
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	lore "github.com/cryptopunkscc/astrald/components/story"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/identifier"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"github.com/cryptopunkscc/astrald/services/util/register"
	"log"
	"time"
)

func init() {
	_ = node.RegisterService(Port, run)
}

const Port = "lore"

const storyMimeType = "application/lore"

var storyMimeTypeBytes = bytes.NewBufferString(storyMimeType).Bytes()

func run(ctx context.Context, core api.Core) error {
	observers := map[api.Stream]string{}
	repository := repo.NewRepoClient(ctx, core)

	go func() {
		time.Sleep(1 * time.Second)

		// Connect to identifier
		mimeTypeSize := byte(len(storyMimeTypeBytes))
		stream, err := connect.Local(ctx, core, identifier.Port, mimeTypeSize)
		if err != nil {
			log.Println(Port, "cannot connect", identifier.Port, err)
			return
		}
		log.Println(Port, "connected to", identifier.Port, err)

		// Send observed type
		_, err = stream.Write(storyMimeTypeBytes)
		if err != nil {
			log.Println(Port, "cannot request observe", identifier.Port, err)
			return
		}
		log.Println(Port, "observing", identifier.Port, err)

		for {
			// Resolve id
			id, _, err := fid.Read(stream)
			if err != nil {
				log.Println(Port, "read new fid from", identifier.Port, err)
				return
			}
			log.Println(Port, "new file fid", id.String())

			go func() {

				// Connecting to repo
				stream, err := repository.Reader(id)
				if err != nil {
					log.Println(Port, "cannot read from", repo.Port, err)
					return
				}

				// Read story
				story, err := lore.Unpack(stream)
				if err != nil {
					log.Println(Port, "cannot unpack story", err)
					return
				}
				storyType := story.Type()
				log.Println(Port, "resolved story type", storyType, "from", repo.Port)

				// Notify observers
				for observer, observerTyp := range observers {
					if storyType == observerTyp {
						err = story.Write(observer)
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
	handler, err := register.Port(ctx, core, Port)
	if err != nil {
		return err
	}
	for request := range handler.Requests() {
		stream := accept.Request(ctx, request)
		log.Println(Port, "accepted new connection")

		go func() {
			defer stream.Close()

			// Read type
			typ, err := stream.ReadStringWithSize()
			if err != nil {
				return
			}

			// Register observer
			observers[stream] = typ
			log.Println(Port, "added new files observer for", typ)

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
