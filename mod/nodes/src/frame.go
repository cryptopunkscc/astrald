package nodes

import "github.com/cryptopunkscc/astrald/mod/nodes/src/frames"

type Frame struct {
	frames.Frame
	Source *Stream
}
