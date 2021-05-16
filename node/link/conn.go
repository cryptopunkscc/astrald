package link

import (
	"github.com/cryptopunkscc/astrald/node/mux"
	"io"
	"sync"
)

// Conn represents an open connection to the remote party's port. Shouldn't be instantiated directly.
type Conn struct {
	io.Reader
	link           *Link
	localStream    mux.Stream
	remoteStreamID mux.StreamID
	mu             sync.Mutex
}

// newConn instantiates a new Conn and starts the necessary routines
func newConn(link *Link, localStream mux.Stream, remoteStreamID mux.StreamID) *Conn {
	var writer io.WriteCloser
	var conn = &Conn{
		link:           link,
		localStream:    localStream,
		remoteStreamID: remoteStreamID,
	}

	conn.Reader, writer = io.Pipe()

	go streamToWriter(localStream, writer)

	return conn
}

// Write writes a byte buffer to the connectiontion
func (conn *Conn) Write(data []byte) (n int, err error) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.link == nil {
		return 0, ErrLinkClosed
	}

	left := data[:]

	for len(left) > 0 {
		chunkLen := mux.MaxPayloadSize
		if chunkLen > len(left) {
			chunkLen = len(left)
		}

		err = conn.link.mux.Write(mux.Frame{
			StreamID: conn.remoteStreamID,
			Data:     left[0:chunkLen],
		})
		if err != nil {
			return
		}

		n += chunkLen
		left = left[chunkLen:]
	}

	return
}

// Close closes the connection
func (conn *Conn) Close() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.link == nil {
		return ErrLinkClosed
	}

	err := conn.link.sendReject(conn.remoteStreamID)
	if err != nil {
		return err
	}

	err = conn.localStream.Close()
	if err != nil {
		return err
	}

	conn.link = nil
	return nil
}

// streamToWriter reads mux frames from a mux stream and writes them to an io.Writer. This allows
// functions that expect an io.Reader interface to read from a mux stream.
func streamToWriter(localStream mux.Stream, w io.WriteCloser) {
	for frame := range localStream.Frames() {
		// An empty frame means close the connection
		if frame.IsEmpty() {
			break
		}

		// Write the data to the pipe
		n, err := w.Write(frame.Data)
		if (err != nil) || (n != len(frame.Data)) {
			break
		}
	}
	w.Close()
}
