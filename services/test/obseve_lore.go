package test

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/story"
	"github.com/cryptopunkscc/astrald/services/lore"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"log"
	"time"
)

func observeLore(ctx context.Context, core api.Core) {
	time.Sleep(1 * time.Second)
	c, err := connect.Local(ctx, core, lore.Port)
	if err != nil {
		return
	}

	_, err = c.WriteStringWithSize8(testStoryType)
	if err != nil {
		return
	}

	log.Println(port, "reading stories")
	for {
		s, err := story.Unpack(c)
		if err != nil {
			log.Println(port, "cannot read story", err)
			return
		}

		log.Println(port, "received story", *s)
	}
}
