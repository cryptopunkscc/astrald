package notify

import (
	"context"
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/enc"
	_node "github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

type CreateChannel struct{ Api }
type DispatchNotification struct{ Api }

func (cc CreateChannel) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(createChannel)
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for query := range port.Queries() {
			conn := query.Accept()
			go func(conn io.ReadWriteCloser) {
				defer conn.Close()
				var channel Channel
				err := gob.NewDecoder(conn).Decode(&channel)
				if err != nil {
					log.Println("Cannot decode notification channel", err)
					return
				}
				err = cc.Create(channel)
				if err != nil {
					log.Println("Cannot create notification channel", err)
					return
				}
				_ = enc.Write(conn, uint8(0))
			}(conn)
		}
	}()

	<-ctx.Done()
	return nil
}

func (dn DispatchNotification) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(notify)
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for query := range port.Queries() {
			conn := query.Accept()
			go func(conn io.ReadWriteCloser) {
				defer conn.Close()
				var notifications []Notification
				err := gob.NewDecoder(conn).Decode(&notifications)
				if err != nil {
					log.Println("Cannot decode notifications", err)
					return
				}
				err = dn.Notify(notifications...)
				if err != nil {
					log.Println("Cannot dispatch notifications", err)
					return
				}
				_ = enc.Write(conn, uint8(0))
			}(conn)
		}
	}()

	<-ctx.Done()
	return nil
}
