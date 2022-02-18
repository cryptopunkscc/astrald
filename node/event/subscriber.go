package event

import "github.com/cryptopunkscc/astrald/sig"

type Subscriber interface {
	Subscribe(cancel sig.Signal) <-chan Event
}
