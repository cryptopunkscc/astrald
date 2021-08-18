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
	"io"
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

		var err error
		var stream api.Stream

		// Connect to identifier
		mimeTypeSize := byte(len(storyMimeTypeBytes))
		if stream, err = connect.Local(ctx, core, identifier.Port, mimeTypeSize); err != nil {
			log.Println(Port, "cannot connect", identifier.Port, err)
			return
		} else {
			log.Println(Port, "connected to", identifier.Port, err)
		}

		// Send observed type
		if _, err = stream.Write(storyMimeTypeBytes); err != nil {
			log.Println(Port, "cannot request observe", identifier.Port, err)
			return
		}

		log.Println(Port, "observing", identifier.Port, err)
		for {
			// Resolve id
			var id fid.ID
			if id, _, err = fid.Read(stream); err != nil {
				log.Println(Port, "read new fid from", identifier.Port, err)
				return
			} else {
				log.Println(Port, "new file fid", id.String())
			}

			go func() {
				var reader io.ReadCloser
				var story *lore.Story
				var storyType string

				// Connecting to repo
				if reader, err = repository.Reader(id); err != nil {
					log.Println(Port, "cannot read from", repo.Port, err)
					return
				}

				// Read story
				defer reader.Close()
				if story, err = lore.Unpack(reader); err != nil {
					log.Println(Port, "cannot unpack story", err)
					return
				}

				// Notify observers
				storyType = story.Type()
				log.Println(Port, "resolved story type", storyType, "from", repo.Port)
				for observer, observerTyp := range observers {
					if storyType == observerTyp {
						if err = story.Write(observer); err != nil {
							log.Println(Port, "cannot send story for", storyType, err)
						}
						log.Println(Port, "sent story of type", storyType)
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
