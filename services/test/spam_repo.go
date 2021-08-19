package test

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/story"
	"github.com/cryptopunkscc/astrald/services/repo"
	"log"
	"time"
)

func spamRepo(ctx context.Context, core api.Core) {
	for {
		time.Sleep(2 * time.Second)

		repository := repo.NewRepoClient(ctx, core)

		log.Println(port, "getting repo writer")
		writer, err := repository.Writer()
		if err != nil {
			log.Println(port, "cannot write repo")
			return
		}

		log.Println(port, "sending story")
		s := story.NewStory(
			time.Now().Unix(),
			testStoryType,
			testStoryAuthor,
			[]fid.ID{},
			[]byte{},
		)
		_, err = writer.WriteWithSize32(s.Pack())
		if err != nil {
			log.Println(port, "cannot send story")
			return
		}

		log.Println(port, "sent story")
		id, err := writer.Finalize()
		if err != nil {
			log.Println(port, "cannot finalize story")
			return
		}

		log.Println(port, "finalized", id.Pack())
	}
}
