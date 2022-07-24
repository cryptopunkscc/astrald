package content

import (
	"context"
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mobile/android/node/android"
	_node "github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

type GetInfo struct{ android.Api }
type Read struct{ android.Api }

func (api GetInfo) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(info)
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for query := range port.Queries() {
			conn, err := query.Accept()
			if err != nil {
				log.Println("Cannot accept query", err)
				continue
			}
			go func(conn io.ReadWriteCloser) {
				defer conn.Close()
				var uri string
				err := cslq.Decode(conn, "[c]c", &uri)
				if err != nil {
					log.Println("Cannot read uri", err)
					return
				}
				var files Info
				err = api.Get(info, uri, &files)
				if err != nil {
					log.Println("Cannot get info", err)
					return
				}
				err = gob.NewEncoder(conn).Encode(files)
				if err != nil {
					log.Println("Cannot encode files info", err)
					return
				}
				var code byte
				_ = cslq.Decode(conn, "c", &code)
			}(conn)
		}
	}()

	<-ctx.Done()
	return nil
}

func (api Read) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(content)
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for query := range port.Queries() {
			conn, err := query.Accept()
			if err != nil {
				log.Println("Cannot accept query", err)
				continue
			}
			go func(conn io.ReadWriteCloser) {
				defer conn.Close()
				var uri string
				err := cslq.Decode(conn, "[c]c", &uri)
				if err != nil {
					log.Println("Cannot read uri", err)
					return
				}
				reader, err := api.Read(content, uri)
				defer reader.Close()
				if err != nil {
					log.Println("Cannot get reader for", uri, err)
					return
				}
				_, err = io.Copy(conn, reader)
				if err != nil {
					log.Println("Cannot copy", uri, err)
					return
				}
			}(conn)
		}
	}()

	<-ctx.Done()
	return nil
}
