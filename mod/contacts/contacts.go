package contacts

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/lib/astral"
	_node "github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
)

const serviceHandle = "contacts"

type Contacts struct{}

type Contact struct {
	Id   string
	Name string
}

func (p Contacts) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(serviceHandle)
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
				var contacts []Contact
				for contact := range node.Contacts.All() {
					contacts = append(contacts, Contact{
						Id:   contact.Identity().String(),
						Name: contact.Alias(),
					})
				}
				_ = json.NewEncoder(conn).Encode(contacts)
			}(conn)
		}
	}()

	<-ctx.Done()
	return nil
}

func (p Contacts) String() string {
	return serviceHandle
}

func Query() (peers []Contact, err error) {
	conn, err := astral.DialName("localnode", serviceHandle)
	if err != nil {
		return
	}
	err = json.NewDecoder(conn).Decode(&peers)
	return
}
