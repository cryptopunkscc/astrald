package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"log"
)

var _ infra.Announcer = &Infra{}

func (i *Infra) Announce(ctx context.Context) error {
	var count int

	for network := range i.Networks() {
		announcer, ok := network.(infra.Announcer)
		if !ok {
			continue
		}

		if err := announcer.Announce(ctx); err != nil {
			log.Println("[announce] error:", err)
		} else {
			count++
		}
	}

	if count == 0 {
		return errors.New("failed to announce on any network")
	}

	return nil
}
