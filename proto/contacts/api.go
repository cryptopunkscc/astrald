package contacts

import "github.com/cryptopunkscc/astrald/node/contacts"

type Service interface {
	Contacts() <-chan *contacts.Contact
}

type Contact struct {
	Id   string
	Name string
}
