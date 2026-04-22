package nodes

import "errors"

// sessionReader is a blocking io.Reader for a session.
// It owns only the blocking loop; ack sending is handled by InputBuffer.onRead.
type sessionReader struct {
	buf *InputBuffer
}

func newSessionReader(buf *InputBuffer) *sessionReader {
	return &sessionReader{buf: buf}
}

func (r *sessionReader) Close() {
	r.buf.Close()
}

func (r *sessionReader) Read(p []byte) (n int, err error) {
	for {
		n, err = r.buf.Read(p)
		if err == nil {
			return
		}
		var empty *ErrBufferEmpty
		if !errors.As(err, &empty) {
			return
		}
		<-empty.Wait()
	}
}
