package lore

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/fs"
	"log"
)

func init() {
	_ = node.RegisterService("lore", run)
}

func run(_ context.Context, core api.Core) error {
	network := core.Network()
	handler, err := core.Network().Register("lore")
	if err != nil {
		return err
	}

	go func() {
		conn, err := network.Connect("", "fs")
		if err != nil {
			log.Println("cannot connect to fs", err)
			return
		}
		var idBuff [40]byte
		writeRequest := append(make([]byte, 0), byte(3))
		_, err = conn.Write(writeRequest)
		if err != nil {
			log.Println("cannot request observe", err)
			return
		}
		for {
			read, err := conn.Read(idBuff[:])
			if err != nil || read < 0 {
				log.Println("read new id from fs", err)
				return
			}
			id := fs.Unpack(idBuff)
			log.Println("new file id", id.String())
			stream, err := network.Connect("", "fs")
			if err != nil {
				log.Println("cannot connect to fs", err)
				return
			}
			writeRequest := append(make([]byte, 0), byte(2))
			write, err := stream.Write(writeRequest[:])
			if err != nil || write < 0 {
				log.Println("cannot request read", err)
				return
			}
			var fileBuff [4096]byte
			for  {
				_, err := stream.Read(fileBuff[:])
				if err != nil {
					return
				}
			}
		}
	}()

	for conn := range handler.Requests() {
		conn.Reject()
	}
	return nil
}
