package content

import (
	"context"
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/enc"
	_node "github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

type GetInfo struct{ Api }
type Read struct{ Api }

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
				files, err := gi.Info(uri)
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

func (r Read) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(read)
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
				reader, err := r.Reader(uri)
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
