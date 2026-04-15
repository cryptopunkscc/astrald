package astral

import (
	"github.com/cryptopunkscc/astrald/sig"
)

const OriginNetwork = "network"
const OriginLocal = "local"

type InFlightQuery struct {
	*Query
	Extra sig.Map[string, any]
}

func Launch(query *Query) *InFlightQuery {
	return &InFlightQuery{
		Query: query,
	}
}

func (q *InFlightQuery) IsNetwork() bool {
	o, ok := q.Extra.Get("origin")
	return ok && (o == OriginNetwork)
}

func (q *InFlightQuery) IsLocal() bool {
	o, _ := q.Extra.Get("origin")
	return o == nil || o == "" || o == OriginLocal
}
