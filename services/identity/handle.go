package identity

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/components/uid"
	"log"
)

func (r Request) Update(_ api.Identity, query string, stream sio.ReadWriteCloser) error {
	log.Println(query, "reading card")
	card, err := uid.ReadCard(stream)
	if err != nil {
		log.Println(query, "cannot read card", err)
		return err
	}
	log.Println(query, "updating card", card)
	err = r.ids.Update(card)
	if err != nil {
		log.Println(query, "cannot update card", err)
		return err
	}
	err = stream.WriteByte(0)
	if err != nil {
		return err
	}
	log.Println(query, "card updated", card)
	return nil
}

func (r Request) List(_ api.Identity, query string, stream sio.ReadWriteCloser) error {
	log.Println(query, "getting cards")
	list, err := r.ids.List()
	if err != nil {
		log.Println(query, "cannot get cards")
		return err
	}
	log.Println(query, "sending size")
	_, err = stream.WriteUInt16(uint16(len(list)))
	if err != nil {
		log.Println(query, "cannot send size")
		return err
	}
	log.Println(query, "sending cards")
	for i, card := range list {
		err = uid.WriteCard(stream, card)
		if err != nil {
			log.Println(query, "cannot send cards", i, err)
			return err
		}
	}
	return nil
}

func (r Request) Get(caller api.Identity, query string, stream sio.ReadWriteCloser) error {
	log.Println(query, "reading id")
	id, err := stream.ReadStringWithSize8()
	if err != nil {
		log.Println(query, "cannot read id", err)
		return err
	}
	log.Println(query, "reading card")
	card, err := r.ids.Get(api.Identity(id))
	if err != nil {
		log.Println(query, "cannot read card", err)
		return err
	}
	log.Println(query, "sending card", err)
	err = uid.WriteCard(stream, *card)
	if err != nil {
		log.Println(query, "cannot send card", err)
		return err
	}
	log.Println(query, "sent  card", err)
	return nil
}