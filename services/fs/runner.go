package fs

import (
	"context"
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

type Runner struct{}

func init() {
	_ = node.RegisterService("fs", run)
}

const (
	RequestRead    = 1
	RequestWrite   = 2
	RequestObserve = 3
)

var AstralHome string

func run(_ context.Context, core api.Core) error {
	observers := map[api.Stream]api.Stream{}
	fs := StorageAdapter(FileStorage(AstralHome))
	handler, err := core.Network().Register("fs")
	if err != nil {
		return err
	}
	for conn := range handler.Requests() {
		stream := conn.Accept()
		caller := conn.Caller()
		log.Println("fs accepted connection")
		go func() {
			defer stream.Close()
			var requestType [1]byte
			l, err := stream.Read(requestType[:])
			if err != nil || l != 1 {
				log.Println("fs error reading type", err)
				return
			}
			log.Println("fs request type", requestType[0])
			switch requestType[0] {
			case RequestRead:
				// Read fileid requested file id
				var idBuff [40]byte
				read, err := stream.Read(idBuff[:])
				if err != nil || read != 40 {
					return
				}
				id := Unpack(idBuff)
				// Obtain file reader
				reader, err := fs.Reader(id)
				if err != nil {
					return
				}
				// Send requested file
				_, err = io.Copy(stream, reader)
			case RequestWrite:
				if caller != core.Network().Identity() {
					return
				}
				var sizeBuff [8]byte
				for {
					// Read next file size
					read, err := stream.Read(sizeBuff[:])
					if err != nil || read != 8 {
						log.Println("closing stream for write request", err)
						return
					}
					log.Println("received bytes size:", sizeBuff)
					size := int64(binary.BigEndian.Uint64(sizeBuff[:]))
					log.Println("parsed size:", size)
					// Obtain file writer
					writer, err := fs.Writer()
					if err != nil {
						log.Println("error while obtaining writer", err)
						return
					}
					log.Println("obtained writer")
					_, err = io.CopyN(writer, stream, size)
					if err != nil {
						log.Println("cannot write to file")
						return
					}
					log.Println("successful write")
					id, err := writer.Finalize()
					if err != nil {
						log.Println("cannot finalize", err)
						return
					}
					packed := id.Pack()
					log.Println("notifying observers")
					for _, observer := range observers {
						_, err := observer.Write(packed[:])
						if err != nil {
							log.Println("cannot notify observer:", err)
						}
					}
					log.Println("sending ")
					_, err = stream.Write(packed[:])
					if err != nil {
						log.Println("cannot write file id", err)
						return
					}
				}
			case RequestObserve:
				if caller != core.Network().Identity() {
					return
				}
				observers[stream] = stream
				log.Println("added new files observer")
				var buffer [1]byte
				for {
					read, err := stream.Read(buffer[:])
					if err != nil || read < 0 {
						log.Println("removing file observer")
						delete(observers, stream)
						return
					}
				}
			}
		}()
	}
	return nil
}
