package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

type Runner struct{}

func init() {
	_ = node.RegisterService("fs", &Runner{})
}

const (
	RequestRead  = 1
	RequestWrite = 2
	RequestObserve = 3
)

var AstralHome string

func (runner *Runner) Run(_ context.Context, core api.Core) error {
	fs := StorageAdapter(FileStorage(AstralHome))
	handler, err := core.Network().Register("fs")
	if err != nil {
		return err
	}
	for conn := range handler.Requests() {
		stream := conn.Accept()
		log.Println("fs accepted connection")
		func() {
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
				var idBuff [40]byte
				read, err := stream.Read(idBuff[:])
				if err != nil || read != 40 {
					return
				}
				id := Unpack(idBuff)
				reader, err := fs.Reader(id)
				if err != nil {
					return
				}
				_, err = io.Copy(stream, reader)
			case RequestWrite:
				if conn.Caller() != core.Network().Identity() {
					return
				}
				writer, err := fs.Writer()
				if err != nil {
					log.Println("error while obtaining writer", err)
					return
				}
				log.Println("obtained writer")
				_, err = io.Copy(writer, stream)
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
				log.Println("file finalized")
				packed := id.Pack()
				_, err = stream.Write(packed[:])
			case RequestObserve:

			}
		}()
	}
	return nil
}
