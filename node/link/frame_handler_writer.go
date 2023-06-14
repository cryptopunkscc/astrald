package link

import (
	"github.com/cryptopunkscc/astrald/mux"
	"io"
)

type WriterFrameHandler struct {
	io.WriteCloser
}

func (f WriterFrameHandler) HandleFrame(frame mux.Frame) (err error) {
	if frame.EOF() {
		err = f.Close()
	} else {
		_, err = f.Write(frame.Data)
	}
	return
}

type CaptureFrameHandler struct {
	mux.FrameHandler
	capture func(mux.Frame)
}

func NewCaptureFrameHandler(frameHandler mux.FrameHandler, capture func(mux.Frame)) *CaptureFrameHandler {
	return &CaptureFrameHandler{FrameHandler: frameHandler, capture: capture}
}

func (h *CaptureFrameHandler) HandleFrame(frame mux.Frame) (err error) {
	if h.capture != nil {
		h.capture(frame)
		h.capture = nil
		return
	}
	return h.FrameHandler.HandleFrame(frame)
}
