package archives

import "github.com/cryptopunkscc/astrald/object"

type EventArchiveIndexed struct {
	ObjectID object.ID
	Archive  *Archive
}
