package messenger

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/story"
	"github.com/cryptopunkscc/astrald/services/messenger/message"
	"github.com/cryptopunkscc/astrald/services/push"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
	"time"
)

func (srv service) Send(
	rc request.Context,
) {
	// Reading a message for send
	log.Println(rc.Port, "reading message for send")
	m, err := message.Read(rc)
	if err != nil {
		log.Println(rc.Port, "cannot read message", err)
		return
	}

	// Getting repository writer
	w, err := srv.Writer()
	if err != nil {
		log.Println(rc.Port, "cannot get writer", err)
		return
	}

	// Preparing story for send
	m.SetSender(srv.Network().Identity())
	s := story.NewStory(
		time.Now().Unix(),
		message.StoryType,
		srv.Network().Identity(),
		[]fid.ID{},
		m.Pack(),
	).Pack()

	// Writing story size to repo writer
	_, err = w.WriteUInt32(uint32(len(s)))
	if err != nil {
		log.Println(rc.Port, "cannot write size of message to repo", err)
		return
	}

	// Writing message to repo
	log.Println(rc.Port, "saving message to repo")
	_, err = w.Write(s)
	if err != nil {
		log.Println(rc.Port, "cannot write message to repo", err)
		return
	}

	// Getting message file id
	log.Println(rc.Port, "getting message id")
	id, err := w.Finalize()
	if err != nil {
		log.Println(rc.Port, "cannot finalize", err)
		return
	}

	// Connecting to remote push service
	log.Println(rc.Port, "connecting to push service", m.Recipient())
	remote, err := connect.RemoteRequest(srv, srv, api.Identity(m.Recipient()), push.Port, push.Push)
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
