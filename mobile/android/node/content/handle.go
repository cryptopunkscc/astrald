package content

import (
	"context"
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/mobile/android/node/android"
	_node "github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

type GetInfo struct{ android.Api }
type Read struct{ android.Api }

func (gi GetInfo) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(info)
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for query := range port.Queries() {
			conn := query.Accept()
			go func(conn io.ReadWriteCloser) {
				defer conn.Close()
				uri, err := enc.ReadL8String(conn)
				if err != nil {
					log.Println("Cannot read uri", err)
					return
				}
				var files Info
				err = gi.Get(info, uri, &files)
				if err != nil {
					log.Println("Cannot get info", err)
					return
				}
				err = gob.NewEncoder(conn).Encode(files)
				if err != nil {
					log.Println("Cannot encode files info", err)
					return
				}
				_, _ = enc.ReadUint8(conn)
			}(conn)
		}
	}()

	<-ctx.Done()
	return nil
}

func (and Read) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(content)
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for query := range port.Queries() {
			conn := query.Accept()
			go func(conn io.ReadWriteCloser) {
				defer conn.Close()
				uri, err := enc.ReadL8String(conn)
				if err != nil {
					log.Println("Cannot read uri", err)
					return
				}
				reader, err := and.Read(content, uri)
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
