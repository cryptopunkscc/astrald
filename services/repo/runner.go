package repo

import (
	"context"
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/api"
	_id "github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/storage/file"
	"github.com/cryptopunkscc/astrald/components/storage/repo"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

func init() {
	_ = node.RegisterService(Port, run)
}

const Port = "repo"

const (
	RequestRead    = 1
	RequestWrite   = 2
	RequestObserve = 3
)

var AstralHome string

func run(ctx context.Context, core api.Core) error {
	observers := map[api.Stream]struct{}{}
	fs := repo.NewAdapter(file.NewStorage(AstralHome))
	handler, err := core.Network().Register(Port)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = handler.Close()
	}()
	for conn := range handler.Requests() {
		if conn.Caller() != core.Network().Identity() {
			conn.Reject()
			log.Println(Port, "rejected remote connection")
			continue
		}
		stream := conn.Accept()
		log.Println(Port, "accepted connection")
		go func() {
			defer stream.Close()
			var requestType [1]byte
			l, err := stream.Read(requestType[:])
			if err != nil || l != 1 {
				log.Println(Port, "error reading type", err)
				return
			}
			log.Println(Port, "request type", requestType[0])
			switch requestType[0] {
			case RequestRead:
				// Read file id requested file fid
				var idBuff [40]byte
				read, err := stream.Read(idBuff[:])
				if err != nil || read != 40 {
					return
				}
				id := _id.Unpack(idBuff)
				// Obtain file reader
				reader, err := fs.Reader(id)
				if err != nil {
					return
				}
				// Send requested file
				_, err = io.Copy(stream, reader)
			case RequestWrite:
				var sizeBuff [8]byte
				for {
					// Read next file size
					read, err := stream.Read(sizeBuff[:])
					if err != nil || read != 8 {
						log.Println(Port, "closing stream for write request", err)
						return
					}
					log.Println(Port, "received bytes size:", sizeBuff)
					size := int64(binary.BigEndian.Uint64(sizeBuff[:]))
					log.Println(Port, "parsed size:", size)
					// Obtain file writer
					writer, err := fs.Writer()
					if err != nil {
						log.Println(Port, "error while obtaining writer", err)
						return
					}
					log.Println(Port, "obtained writer")
					_, err = io.CopyN(writer, stream, size)
					if err != nil {
						log.Println(Port, "cannot write to file")
						return
					}
					log.Println(Port, "successful write")
					id, err := writer.Finalize()
					if err != nil {
						log.Println(Port, "cannot finalize", err)
						return
					}
					packed := id.Pack()
					log.Println(Port, "notifying observers", len(observers))
					for observer := range observers {
						_, err := observer.Write(packed[:])
						if err != nil {
							log.Println(Port, "cannot notify observer:", err)
						}
					}
					log.Println(Port, "sending")
					_, err = stream.Write(packed[:])
					if err != nil {
						log.Println(Port, "cannot write file fid", err)
						return
					}
				}
			case RequestObserve:
				observers[stream] = struct{}{}
				log.Println(Port, "added new files observer")
				var buffer [1]byte
				for {
					_, err := stream.Read(buffer[:])
					if err != nil {
						log.Println(Port, "removing file observer", err)
						delete(observers, stream)
						return
					}
				}
			}
		}()
	}
	return nil
}
