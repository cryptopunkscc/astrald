package handle

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/components/story"
	"github.com/cryptopunkscc/astrald/services/messenger/message"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"log"
	"time"
)

func Send(
	ctx context.Context,
	core api.Core,
	r repo.LocalRepository,
	stream sio.ReadWrite,
	Port string,
) {
	log.Println(Port, "reading for send message")
	m, err := message.Read(stream)
	if err != nil {
		log.Println(Port, "cannot read message", err)
		return
	}
	w, err := r.Writer()
	if err != nil {
		log.Println(Port, "cannot get writer", err)
		return
	}
	s := story.NewStory(
		time.Now().Unix(),
		message.StoryType,
		core.Network().Identity(),
		[]fid.ID{},
		m.Pack(),
	)
	log.Println(Port, "saving message to repo")
	err = s.Write(w)
	if err != nil {
		log.Println(Port, "cannot write message to repo", err)
		return
	}
	log.Println(Port, "getting message id", err)
	id, err := w.Finalize()
	if err != nil {
		log.Println(Port, "cannot finalize", err)
		return
	}
	log.Println(Port, "connecting to ", m.Recipient())
	remote, err := connect.Remote(ctx, core, api.Identity(m.Recipient()), Port)
	if err != nil {
		log.Println(Port, "cannot connect to", m.Recipient(), err)
		return
	}
	defer func() { _ = remote.Close() }()
	err = id.Write(remote)
	if err != nil {
		log.Println(Port, "cannot send message id to", m.Recipient(), err)
		return
	}
}
