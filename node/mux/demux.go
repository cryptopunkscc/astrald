package mux

import (
	"errors"
	"log"
	"math"
	"sync"
)

const defaultStreamID = 0

// Demux reads mux frames and splits them into separate streams
type Demux struct {
	mux     *Mux
	streams map[int]*InputStream
	mu      sync.Mutex
}

func NewDemux(mux *Mux) *Demux {
	demux := &Demux{
		mux:     mux,
		streams: make(map[int]*InputStream),
	}

	demux.Stream() // setup the default stream with id 0

	go func() {
		err := demux.readFrames()
		if err != nil {
		}

		demux.mu.Lock()
		for _, stream := range demux.streams {
			err := stream.Close()
			if err != nil {
			}
		}
		demux.streams = nil
		demux.mu.Unlock()
	}()

	return demux
}

func (demux *Demux) DefaultStream() *InputStream {
	return demux.streams[defaultStreamID]
}

// Stream allocates and returns a new input stream
func (demux *Demux) Stream() (*InputStream, error) {
	demux.mu.Lock()
	defer demux.mu.Unlock()

	if demux.streams == nil {
		return nil, errors.New("demux error")
	}

	for i := 0; i < MaxStreams; i++ {
		if _, used := demux.streams[i]; !used {
			stream := newInputStream(i)

			demux.streams[i] = stream
			go func() {
				<-stream.WaitClose()
				demux.freeInputStream(i)
			}()

			return stream, nil
		}
	}
	return nil, errors.New("all streams used")
}

// freeInputStream releases a wire so that it can be used again
func (demux *Demux) freeInputStream(id int) error {
	demux.mu.Lock()
	defer demux.mu.Unlock()

	if _, used := demux.streams[id]; !used {
		return errors.New("stream not used")
	}

	delete(demux.streams, id)

	return nil
}

func (demux *Demux) readFrames() error {
	buf := make([]byte, math.MaxUint16)

	for {
		// Read next frame from the mux
		localStreamID, payloadLen, err := demux.mux.Read(buf)
		if err != nil {
			return err
		}

		stream, ok := demux.streams[localStreamID]
		if !ok {
			log.Println("[demux] warning: received mux frame on a closed stream")
			continue
		}

		if payloadLen == 0 {
			stream.Close()
		} else {
			stream.write(buf[:payloadLen])
		}
	}
}
