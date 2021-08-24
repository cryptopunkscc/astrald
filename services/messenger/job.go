package messenger

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/components/story"
	"github.com/cryptopunkscc/astrald/services/lore"
	"github.com/cryptopunkscc/astrald/services/messenger/message"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"log"
	"time"
)

func ObserveLore(
	ctx context.Context,
	core api.Core,
	port string,
	observers map[sio.ReadWriteCloser]struct{},
) {
	time.Sleep(2 * time.Second)
	stream, err := connect.Local(ctx, core, lore.Port)
	if err != nil {
		log.Println(port, "cannot connect to", lore.Port)
		return
	}

	log.Println(port, "requesting observe", message.StoryType, lore.Port)
	_, err = stream.WriteStringWithSize8(message.StoryType)
	if err != nil {
		log.Println(port, "cannot observe", message.StoryType, lore.Port)
		return
	}

	log.Println(port, "observing", lore.Port)
	for {
		s, err := story.Unpack(stream)
		if err != nil {
			log.Println(port, "cannot read new fid from", lore.Port, err)
			return
		}

		go func() {
			// Notify observers
			log.Println(port, "notifying observers")
			for observer := range observers {
				if err = s.Write(observer); err != nil {
					log.Println(port, "cannot send message", err)
				}
			}
			log.Println(port, "finish notifying observers")
		}()
	}
}
