package muxlink

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mux"
	"io"
)

type ResponseHandler struct {
	Func func(response Response, err error)
}

func (h *ResponseHandler) HandleMux(event mux.Event) {
	switch event := event.(type) {
	case mux.Frame:
		h.HandleFrame(event)
	}
}

func (h *ResponseHandler) HandleFrame(frame mux.Frame) {
	if frame.IsEmpty() {
		h.Func(Response{}, io.EOF)
		return
	}

	var res Response
	var r = bytes.NewReader(frame.Data)
	var err = cslq.Decode(r, "v", &res)

	h.Func(res, err)
}
