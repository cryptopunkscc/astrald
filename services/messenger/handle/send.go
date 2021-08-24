package handle

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/components/story"
	"github.com/cryptopunkscc/astrald/services/messenger/message"
	"github.com/cryptopunkscc/astrald/services/push"
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
	// Reading a message for send
	log.Println(Port, "reading message for send")
	m, err := message.Read(stream)
	if err != nil {
		log.Println(Port, "cannot read message", err)
		return
	}

	// Getting repository writer
	w, err := r.Writer()
	if err != nil {
		log.Println(Port, "cannot get writer", err)
		return
	}

	// Preparing story for send
	m.SetSender(core.Network().Identity())
	s := story.NewStory(
		time.Now().Unix(),
		message.StoryType,
		core.Network().Identity(),
		[]fid.ID{},
		m.Pack(),
	).Pack()

	// Writing story size to repo writer
	_, err = w.WriteUInt32(uint32(len(s)))
	if err != nil {
		log.Println(Port, "cannot write size of message to repo", err)
		return
	}

	// Writing message to repo
	log.Println(Port, "saving message to repo")
	_, err = w.Write(s)
	if err != nil {
		log.Println(Port, "cannot write message to repo", err)
		return
	}

	// Getting message file id
	log.Println(Port, "getting message id")
	id, err := w.Finalize()
	if err != nil {
		log.Println(Port, "cannot finalize", err)
		return
	}

	// Connecting to remote push service
	log.Println(Port, "connecting to push service", m.Recipient())
	remote, err := connect.RemoteRequest(ctx, core, api.Identity(m.Recipient()), push.Port, push.Push)
	if err != nil {
		log.Println(Port, "cannot connect to", m.Recipient(), err)
		return
	}
	defer func() { _ = remote.Close() }()

	// Sending message file id to remote push service
	log.Println(Port, "sending message id to", m.Recipient(), push.Port)
	err = id.Write(remote)
	if err != nil {
		log.Println(Port, "cannot send message id to", m.Recipient(), err)
		return
	}

	// Wait for response
	log.Println(Port, "waiting for response from", m.Recipient(), push.Port)
	_, err = remote.ReadByte()
	if err != nil {
		log.Println(Port, "error while waiting for ok", m.Recipient())
	}
	log.Println("sending message done")
}
