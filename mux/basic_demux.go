package mux

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"sync"
)

// BasicDemux reads mux frames and splits them into separate streams
type BasicDemux struct {
	reader io.Reader
	readMu sync.Mutex
}

func NewBasicDemux(reader io.Reader) *BasicDemux {
	demux := &BasicDemux{
		reader: reader,
	}

	return demux
}

// ReadFrame reads the next frame and returns stream id, bytes read and an error if one occured
func (dem *BasicDemux) ReadFrame() (streamID int, buf []byte, err error) {
	dem.readMu.Lock()
	defer dem.readMu.Unlock()

	return streamID, buf, cslq.Decode(dem.reader, "s[s]c", &streamID, &buf)
}
