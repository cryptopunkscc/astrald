package files

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/serialize"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/repo"
	"io"
	"log"
)

func init() {
	_ = node.RegisterService(Port, run)
}

const Port = "files"

const (
	RequestRead = 1
	RequestList = 2
)

func run(ctx context.Context, core api.Core) (err error) {
	handler, err := core.Network().Register(Port)
	if err != nil {
		return
	}
	r := repo.NewProxy("", ctx, core)
	for conn := range handler.Requests() {
		stream := conn.Accept()
		log.Println(Port, "accepted connection")

		go func() {
			defer stream.Close()
			s := serialize.NewSerializer(stream)
			request, err := s.ReadByte()
			if err != nil {
				log.Println(Port, "cannot read request type", err)
				return
			}

			switch request {
			case RequestRead:
				for {
					id, err := fid.Resolve(stream)
					if err != nil {
						return
					}
					reader, err := r.Reader(id)
					if err != nil {
						log.Println(Port, "cannot read file from", repo.Port)
						return
					}
					_, err = io.Copy(stream, reader)
					if err != nil {
						log.Println(Port, "cannot send file")
						return
					}
				}
			case RequestList:
				reader, err := r.List()
				if err != nil {
					log.Println(Port, "cannot list files from", repo.Port)
					return
				}
				_, err = io.Copy(stream, reader)
				if err != nil {
					log.Println(Port, "cannot send files")
					return
				}
			}
		}()
	}
	return
}
