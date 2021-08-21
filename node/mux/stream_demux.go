package mux

import (
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
)

const defaultStreamID = 0

// StreamDemux reads mux frames and splits them into separate streams
type StreamDemux struct {
	streams   map[int]*InputStream
	streamsMu sync.Mutex
	rawDemux  *BasicDemux
}

func NewStreamDemux(reader io.Reader) *StreamDemux {
	demux := &StreamDemux{
		rawDemux: NewBasicDemux(reader),
		streams:  make(map[int]*InputStream),
	}

	demux.streams[defaultStreamID] = newInputStream(defaultStreamID)

	go func() {
		err := demux.processFrames()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				panic(err)
			}
		}

		demux.streamsMu.Lock()
		for _, stream := range demux.streams {
			_ = stream.close()
		}
		demux.streams = nil
		demux.streamsMu.Unlock()
	}()

	return demux
}

func (demux *StreamDemux) DefaultStream() *InputStream {
	return demux.streams[defaultStreamID]
}

// Stream allocates and returns a new input stream
func (demux *StreamDemux) Stream() (*InputStream, error) {
	demux.streamsMu.Lock()
	defer demux.streamsMu.Unlock()

	if demux.streams == nil {
		return nil, errors.New("demux error")
	}

	for i := 0; i < MaxStreams; i++ {
		if _, used := demux.streams[i]; !used {
			stream := newInputStream(i)

			demux.streams[i] = stream
			go func() {
				<-stream.WaitClose()
				demux.removeInputStream(i)
			}()

			return stream, nil
		}
	}
	return nil, errors.New("all streams used")
}

// removeInputStream releases a stream so that it can be used again
func (demux *StreamDemux) removeInputStream(id int) error {
	demux.streamsMu.Lock()
	defer demux.streamsMu.Unlock()

	if _, used := demux.streams[id]; !used {
		return errors.New("stream not used")
	}

	delete(demux.streams, id)

	return nil
}

func (demux *StreamDemux) processFrames() error {
	buf := make([]byte, math.MaxUint16)

	for {
		// Read next frame from the mux
		localStreamID, payloadLen, err := demux.rawDemux.ReadFrame(buf)
		if err != nil {
			return err
		}

		stream, ok := demux.streams[localStreamID]
		if !ok {
			return fmt.Errorf("stream %d is closed", localStreamID)
		}

		if payloadLen == 0 {
			stream.close()
		} else {
			stream.write(buf[:payloadLen])
		}
	}
}
