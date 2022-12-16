package mux

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"sync"
)

// FrameDemux reads frames from the multiplexed stream
type FrameDemux struct {
	reader io.Reader
	readMu sync.Mutex
}

type Frame struct {
	StreamID int
	Data     []byte
}

// NewFrameDemux returns a new FrameDemux that reads frames from the provided reader
func NewFrameDemux(reader io.Reader) *FrameDemux {
	demux := &FrameDemux{
		reader: reader,
	}

	return demux
}

// ReadFrame reads the next frame and returns stream id, bytes read and an error if one occured
func (dem *FrameDemux) ReadFrame() (frame Frame, err error) {
	dem.readMu.Lock()
	defer dem.readMu.Unlock()

	return frame, cslq.Decode(dem.reader, "s[s]c", &frame.StreamID, &frame.Data)
}

func (f Frame) Size() int {
	return len(f.Data)
}
