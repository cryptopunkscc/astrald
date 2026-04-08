package apps

import (
	"io"
	"sync"
)

type lockableWriteCloser struct {
	io.WriteCloser
	sync.Mutex
}

func newLockableWriteCloser(writeCloser io.WriteCloser) *lockableWriteCloser {
	return &lockableWriteCloser{WriteCloser: writeCloser}
}

func (s *lockableWriteCloser) Close() error {
	s.Lock()
	defer s.Unlock()

	return s.WriteCloser.Close()
}

func (s *lockableWriteCloser) Write(p []byte) (n int, err error) {
	s.Lock()
	defer s.Unlock()

	return s.WriteCloser.Write(p)
}
