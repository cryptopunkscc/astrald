package frames

import "github.com/cryptopunkscc/astrald/astral"

var FrameTypes = []string{
	"nodes.frames.ping",
	"nodes.frames.query",
	"nodes.frames.relay_query",
	"nodes.frames.read",
	"nodes.frames.response",
	"nodes.frames.data",
	"nodes.frames.migrate",
	"nodes.frames.reset",
}

var FrameTypeEncoder = astral.IndexedTypeEncoder(FrameTypes)
var FrameTypeDecoder = astral.IndexedTypeDecoder(FrameTypes)
