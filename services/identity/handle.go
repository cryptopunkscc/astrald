package identity

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/uid"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

func (srv *service) Update(rc request.Context) error {
	log.Println(rc.Port, "reading card")
	card, err := uid.ReadCard(rc)
	if err != nil {
		log.Println(rc.Port, "cannot read card", err)
		return err
	}
	log.Println(rc.Port, "updating card", card)
	err = srv.ids.Update(card)
	if err != nil {
		log.Println(rc.Port, "cannot update card", err)
		return err
	}
	err = rc.WriteByte(0)
	if err != nil {
		return err
	}
	log.Println(rc.Port, "card updated", card)
	return nil
}

func (srv *service) List(rc request.Context) error {
	log.Println(rc.Port, "getting cards")
	list, err := srv.ids.List()
	if err != nil {
		log.Println(rc.Port, "cannot get cards")
		return err
	}
	log.Println(rc.Port, "sending size")
	_, err = rc.WriteUInt16(uint16(len(list)))
	if err != nil {
		log.Println(rc.Port, "cannot send size")
		return err
	}
	log.Println(rc.Port, "sending cards")
	for i, card := range list {
		err = uid.WriteCard(rc, card)
		if err != nil {
			log.Println(rc.Port, "cannot send cards", i, err)
			return err
		}
	}
	return nil
}

func (srv *service) Get(rc request.Context) error {
	log.Println(rc.Port, "reading id")
	id, err := rc.ReadStringWithSize8()
	if err != nil {
		log.Println(rc.Port, "cannot read id", err)
		return err
	}
	log.Println(rc.Port, "reading card")
	card, err := srv.ids.Get(api.Identity(id))
	if err != nil {
		log.Println(rc.Port, "cannot read card", err)
		return err
	}
	log.Println(rc.Port, "sending card", err)
	err = uid.WriteCard(rc, *card)
	if err != nil {
		log.Println(rc.Port, "cannot send card", err)
		return err
	}
	log.Println(rc.Port, "sent  card", err)
	return nil
}