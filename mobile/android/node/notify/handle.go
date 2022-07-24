package notify

import (
	"context"
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mobile/android/node/android"
	_node "github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

type CreateChannel struct{ android.Api }
type DispatchNotification struct{ android.Api }

func (api CreateChannel) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(createChannel)
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
				var channel Channel
				err := gob.NewDecoder(conn).Decode(&channel)
				if err != nil {
					log.Println("Cannot decode notification channel", err)
					return
				}
				err = api.Call(createChannel, channel)
				if err != nil {
					log.Println("Cannot create notification channel", err)
					return
				}
				_ = cslq.Encode(conn, "c", 0)
			}(conn)
		}
	}()

	<-ctx.Done()
	return nil
}

func (api DispatchNotification) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(notify)
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
				decoder := gob.NewDecoder(conn)
				for {
					var notifications []Notification
					err := decoder.Decode(&notifications)
					if err != nil {
						if err != io.EOF {
							log.Println("Cannot decode notifications", err)
						}
						return
					}
					err = api.Call(notify, notifications)
					if err != nil {
						log.Println("Cannot dispatch notifications", err)
						//return
					}
					_ = cslq.Encode(conn, "c", 0)
				}
			}(conn)
		}
	}()

	<-ctx.Done()
	return nil
}
