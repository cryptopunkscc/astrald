package mux

import (
	"io"
	"log"
	"sync"
)

type StreamID uint16

// Mux is a simple multiplexer that can exchange frames over the provided transport. It supports up to 65535 concurrent
// streams since it uses a 16-bit int for stream addressing.
type Mux struct {
	transport io.ReadWriteCloser
	dataMu    sync.Mutex
	dataCh    map[StreamID]chan Frame // data channels
	controlCh chan Frame              // control channel
	error     error                   // after link closes this contains the error or nil //TODO: proper error handling
}

// New instantiates a new multiplexer over the provided transport
func New(transport io.ReadWriteCloser) *Mux {
	mux := &Mux{
		transport: transport,
		dataCh:    make(map[StreamID]chan Frame),
		controlCh: make(chan Frame),
	}

	// Start processing the incoming frames
	go mux.process()

	return mux
}

// Control returns a channel for reading control frames
func (mux *Mux) Control() <-chan Frame {
	return mux.controlCh
}

// Stream reserves a local stream and returns a reader/closer for the stream
func (mux *Mux) Stream() (Stream, error) {
	mux.dataMu.Lock()
	defer mux.dataMu.Unlock()

	// find a free stream
	for id := MinStreamID; id < MaxStreamID; id++ {
		if _, used := mux.dataCh[id]; used {
			continue
		}

		ch := make(chan Frame)
		mux.dataCh[id] = ch
		return Stream{
			id:     id,
			frames: ch,
			mux:    mux,
		}, nil
	}

	return Stream{}, ErrTooManyStreams
}

// Write a frame to the transport
func (mux *Mux) Write(frame Frame) error {
	return writeFrame(mux.transport, uint16(frame.StreamID), frame.Data)
}

// Error returns the error state of the multiplexer
func (mux *Mux) Error() error {
	return mux.error
}

// process mux frames coming from the other party
func (mux *Mux) process() {
	for {
		frame, err := mux.readFrame()
		if err != nil {
			mux.setError(err)
			break
		}

		if frame.StreamID == ControlStreamID {
			err = mux.handleControl(frame)
			if err != nil {
				log.Println("error handling mux ctl frame:", err)
			}
		} else {
			err = mux.handleData(frame)
			if err != nil {
				log.Println("error handling mux data frame:", err)
			}
		}
	}

	// Clean up
	mux.closeAllChannels()
	close(mux.controlCh)
}

// readFrame reads and returns the next Frame from the transport
func (mux *Mux) readFrame() (Frame, error) {
	ch, data, err := readFrame(mux.transport)
	if err != nil {
		return Frame{}, err
	}

	return Frame{
		StreamID: StreamID(ch),
		Data:     data,
	}, nil
}

// closeChannel closes local channel and frees it for reuse
func (mux *Mux) closeChannel(id StreamID) error {
	mux.dataMu.Lock()
	defer mux.dataMu.Unlock()

	ch, found := mux.dataCh[id]
	if !found {
		return ErrStreamNotFound
	}

	close(ch)
	delete(mux.dataCh, id)
	return nil
}

// closeAllChannels closes all data channels.
func (mux *Mux) closeAllChannels() {
	mux.dataMu.Lock()
	defer mux.dataMu.Unlock()

	for id, ch := range mux.dataCh {
		close(ch)
		delete(mux.dataCh, id)
	}
}

// handleControl sends a control frame to the control channel
func (mux *Mux) handleControl(frame Frame) error {
	mux.controlCh <- frame
	return nil
}

// handleData routes the data frame to the respective channel
func (mux *Mux) handleData(frame Frame) error {
	mux.dataMu.Lock()
	defer mux.dataMu.Unlock()

	id := frame.StreamID
	ch, exists := mux.dataCh[id]
	if !exists {
		return ErrStreamNotFound
	}
	ch <- frame
	return nil
}

// setError sets the multiplexer error field
func (mux *Mux) setError(err error) {
	mux.error = err
}
