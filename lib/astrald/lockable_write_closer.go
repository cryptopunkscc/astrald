package astrald

import (
	"io"
	"sync"
)

// LockableWriteCloser wraps an io.WriteCloser and provides a lock on the Write and Close methods.
type LockableWriteCloser struct {
	io.WriteCloser
	sync.Mutex
}

func NewLockableWriteCloser(writeCloser io.WriteCloser) *LockableWriteCloser {
	return &LockableWriteCloser{WriteCloser: writeCloser}
}

func (s *LockableWriteCloser) Close() error {
	s.Lock()
	defer s.Unlock()

	return s.WriteCloser.Close()
}

func (s *LockableWriteCloser) Write(p []byte) (n int, err error) {
	s.Lock()
	defer s.Unlock()

	return s.WriteCloser.Write(p)
}
