package lore

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/serialize"
	_story "github.com/cryptopunkscc/astrald/components/story"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/lore"
	"github.com/cryptopunkscc/astrald/services/repo"
	"log"
	"time"
)

const port = "lore-test"
const testStoryType = "test_type"
const TestStoryAuthor = "test_author"

func init() {
	_ = node.RegisterService(port, run)
}

func run(ctx context.Context, core api.Core) (err error) {
	go func() {
		time.Sleep(1 * time.Second)
		stream, err := core.Network().Connect("", lore.Port)
		if err != nil {
			return
		}

		go func() {
			<-ctx.Done()
			_ = stream.Close()
		}()

		s := serialize.NewSerializer(stream)

		_, err = s.WriteStringWithSize(testStoryType)
		if err != nil {
			return
		}

		for {
			story, err := _story.Unpack(stream)
			if err != nil {
				log.Println(port, "cannot read story", err)
				return
			}

			log.Println(port, "received story", *story)
		}
	}()
	go func() {
		for {
			time.Sleep(2 * time.Second)

			stream, err := core.Network().Connect("", repo.Port)
			if err != nil {
				return
			}
			log.Println(port, "connected to", repo.Port)

			s := serialize.NewSerializer(stream)

			story := _story.NewStory(
				time.Now().Unix(),
				testStoryType,
				TestStoryAuthor,
				[]fid.ID{},
				[]byte{},
			)
			storyBuff := story.PackBytes()
			log.Println(port, "sending story with size", len(storyBuff), storyBuff)
			err = s.WriteByte(repo.RequestWrite)
			if err != nil {
				log.Println(err)
				return
			}
			_, err = s.WriteUInt64(uint64(len(storyBuff)))
			if err != nil {
				log.Println(err)
				return
			}
			_, err = s.Write(storyBuff)
			if err != nil {
				log.Println(err)
				return
			}
			_, err = s.ReadN(fid.Size)
			if err != nil {
				return
			}
			err = stream.Close()
			if err != nil {
				return
			}
		}
	}()

	<-ctx.Done()
	return nil
}