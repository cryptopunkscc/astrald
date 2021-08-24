package lore

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/sio"
	lore "github.com/cryptopunkscc/astrald/components/story"
	"github.com/cryptopunkscc/astrald/services/identifier"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"io"
	"log"
	"time"
)

func (srv service) observeIdentifier() {
	time.Sleep(1 * time.Second)

	var err error
	var stream sio.ReadWriteCloser

	// Connect to identifier
	if stream, err = connect.Local(srv.ctx, srv.core, identifier.Port); err != nil {
		log.Println(Port, "cannot connect", identifier.Port, err)
		return
	}

	// Send observed type
	log.Println(Port, "requesting observe", storyMimeType, identifier.Port, err)
	if _, err = stream.WriteStringWithSize8(storyMimeType); err != nil {
		log.Println(Port, "cannot request observe", identifier.Port, err)
		return
	}

	log.Println(Port, "observing", identifier.Port)
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
			if reader, err = srv.repository.Reader(id); err != nil {
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
			for observer, observerTyp := range srv.observers {
				if storyType == observerTyp {
					if err = story.Write(observer); err != nil {
						log.Println(Port, "cannot send story for", storyType, err)
					}
					log.Println(Port, "sent story of type", storyType)
				}
			}
		}()
	}
}
