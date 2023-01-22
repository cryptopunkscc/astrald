package mux

import (
	"errors"
	"io"
	"log"
	"strings"
	"sync"
)

// StreamDemux reads mux frames and splits them into separate streams
type StreamDemux struct {
	streams      map[int]*InputStream
	streamsMu    sync.Mutex
	frames       *FrameDemux
	lastStreamID int
}

func NewStreamDemux(reader io.Reader) *StreamDemux {
	demux := &StreamDemux{
		frames:  NewFrameDemux(reader),
		streams: make(map[int]*InputStream),
	}

	demux.streams[controlStreamID] = newInputStream(demux, controlStreamID)

	go demux.run()

	return demux
}

// ControlStream returns the default demux stream. Default stream is open as long as the demux session is running.
func (demux *StreamDemux) ControlStream() (*InputStream, error) {
	demux.streamsMu.Lock()
	defer demux.streamsMu.Unlock()

	if demux.streams == nil {
		return nil, ErrDemuxClosed
	}

	return demux.streams[controlStreamID], nil
}

// AllocStream allocates and returns a new input stream
func (demux *StreamDemux) AllocStream() (*InputStream, error) {
	demux.streamsMu.Lock()
	defer demux.streamsMu.Unlock()

	if demux.streams == nil {
		return nil, ErrDemuxClosed
	}

	for i := 0; i < MaxStreams; i++ {
		s := (demux.lastStreamID + i + 1) % MaxStreams
		if _, used := demux.streams[s]; !used {
			stream := newInputStream(demux, s)

			demux.streams[s] = stream
			demux.lastStreamID = s

			return stream, nil
		}
	}
	return nil, ErrStreamLimitReached
}

func (demux *StreamDemux) run() {
	err := demux.processFrames()

	switch {
	case strings.Contains(err.Error(), "use of closed network connection"):
	case strings.Contains(err.Error(), "connection abort"):
	case errors.Is(err, io.EOF):
	case errors.Is(err, errControlStreamClosed):
	default:
		log.Println("mux.StreamDemux.run() error:", err)
	}
}

func (demux *StreamDemux) processFrames() error {
	defer demux.closeDemux()

	for {
		// Read next frame from the demux
		frame, err := demux.frames.ReadFrame()
		if err != nil {
			return err
		}

		stream, ok := demux.getStream(frame.StreamID)
		if !ok {
			return ErrFrameOnClosedStream{frame.StreamID}
		}

		if frame.Size() > 0 {
			// data frame
			stream.write(frame.Data)
		} else {
			// close frame
			if frame.StreamID == controlStreamID {
				return errControlStreamClosed
			}
			stream.closeWriter(io.EOF)
			demux.removeInputStream(stream.id)
		}
	}
}

func (demux *StreamDemux) closeDemux() {
	demux.streamsMu.Lock()
	defer demux.streamsMu.Unlock()

	for _, stream := range demux.streams {
		stream.closeWriter(io.EOF)
		delete(demux.streams, stream.id)
	}
	demux.streams = nil
}

func (demux *StreamDemux) getStream(streamID int) (*InputStream, bool) {
	demux.streamsMu.Lock()
	defer demux.streamsMu.Unlock()

	stream, found := demux.streams[streamID]
	return stream, found
}

// removeInputStream releases a stream so that it can be used again
func (demux *StreamDemux) removeInputStream(id int) error {
	demux.streamsMu.Lock()
	defer demux.streamsMu.Unlock()

	if _, used := demux.streams[id]; !used {
		return ErrInvalidStreamID
	}

	delete(demux.streams, id)

	return nil
}
