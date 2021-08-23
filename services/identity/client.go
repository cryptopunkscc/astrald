package identity

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/uid"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"log"
)

type client struct {
	ctx  context.Context
	core api.Core
}

func NewClient(ctx context.Context, core api.Core) uid.Identities {
	return &client{ctx: ctx, core: core}
}

func (c *client) Update(card uid.Card) error {
	s, err := connect.LocalRequest(c.ctx, c.core, Port, Update)
	if err != nil {
		return err
	}
	err = uid.WriteCard(s, card)
	if err != nil {
		return err
	}
	_, err = s.ReadByte()
	return err
}

func (c *client) List() ([]uid.Card, error) {
	s, err := connect.LocalRequest(c.ctx, c.core, Port, List)
	if err != nil {
		log.Println("cannot request list", err)
		return nil, err
	}
	l, err := s.ReadUint16()
	if err != nil {
		log.Println("cannot read list size", err)
		return nil, err
	}
	cards := make([]uid.Card, l)
	for i := uint16(0); i < l; i++ {
		card, err := uid.ReadCard(s)
		if err != nil {
			log.Println("cannot read card", err)
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func (c *client) Get(id api.Identity) (*uid.Card, error) {
	s, err := connect.LocalRequest(c.ctx, c.core, Port, Get)
	if err != nil {
		return nil, err
	}
	_, err = s.WriteStringWithSize8(string(id))
	if err != nil {
		return nil, err
	}
	card, err := uid.ReadCard(s)
	if err != nil {
		return nil, err
	}
	return &card, err
}