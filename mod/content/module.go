package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

const ModuleName = "content"
const DBPrefix = "content__"

type Module interface {
	Identify(*object.ID) (*TypeInfo, error)
	Forget(*object.ID) error
	Scan(ctx context.Context, opts *ScanOpts) <-chan *TypeInfo

	Ready(ctx context.Context) error
}

type ScanOpts struct {
	Type  string
	After time.Time
}

type TypeInfo struct {
	ObjectID     *object.ID
	Type         astral.String8 // detected data type
	Method       astral.String8 // method used to detect type (adc | mimetype)
	IdentifiedAt astral.Time
}
