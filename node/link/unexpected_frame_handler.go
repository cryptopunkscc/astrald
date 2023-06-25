package link

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/mux"
)

type UnexpectedFrameHandler struct {
	*CoreLink
}

func (handler *UnexpectedFrameHandler) HandleFrame(frame mux.Frame) {
	handler.CloseWithError(
		fmt.Errorf("protocol error: unexpected frame on port %v (len=%v)",
			frame.Port,
			len(frame.Data),
		))
}
