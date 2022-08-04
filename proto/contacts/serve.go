package contacts

import (
	"encoding/json"
	"io"
)

func Serve(s Service, conn io.ReadWriteCloser) error {
	var contacts []Contact
	for contact := range s.Contacts() {
		contacts = append(contacts, Contact{
			Id:   contact.Identity().String(),
			Name: contact.Alias(),
		})
	}
	return json.NewEncoder(conn).Encode(contacts)
}
