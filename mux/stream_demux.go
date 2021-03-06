package mux

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
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
			errClosed := strings.Contains(err.Error(), "use of closed network connection")
			errEOF := errors.Is(err, io.EOF)

			if !(errEOF || errClosed) {
				log.Println("error: demux.processFrames():", err)
			}
		}

		demux.streamsMu.Lock()
		for _, stream := range demux.streams {
			_ = stream.Close()
		}
		demux.streams = nil
		demux.streamsMu.Unlock()
	}()

	return demux
}

func (demux *StreamDemux) DefaultStream() *InputStream {
	demux.streamsMu.Lock()
	defer demux.streamsMu.Unlock()

	return demux.streams[defaultStreamID]
}

// Stream allocates and returns a new input stream
func (demux *StreamDemux) Stream() (*InputStream, error) {
	demux.streamsMu.Lock()
	defer demux.streamsMu.Unlock()

	if demux.streams == nil {
		return nil, errors.New("demux error: demux closed")
	}

	for i := 0; i < MaxStreams; i++ {
		if _, used := demux.streams[i]; !used {
			stream := newInputStream(i)

			demux.streams[i] = stream
			go func() {
				<-stream.Wait()
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
	for {
		// Read next frame from the mux
		localStreamID, buf, err := demux.rawDemux.ReadFrame()
		if err != nil {
			return err
		}

		demux.streamsMu.Lock()
		stream, ok := demux.streams[localStreamID]
		demux.streamsMu.Unlock()
		if !ok {
			return fmt.Errorf("stream %d is closed", localStreamID)
		}

		if len(buf) == 0 {
			stream.Close()
		} else {
			stream.write(buf)
		}
	}
}
