package nodes

import (
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

type Frame struct {
	frames.Frame
	Source *Link
}
