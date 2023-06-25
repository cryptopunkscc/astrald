package link

import (
	"sync"
)

type remoteBuffers struct {
	link  *CoreLink
	sizes map[int]int
	cond  *sync.Cond
}

func newRemoteBuffers(link *CoreLink) *remoteBuffers {
	return &remoteBuffers{
		sizes: map[int]int{},
		cond:  sync.NewCond(&sync.Mutex{}),
		link:  link,
	}
}

func (buffers *remoteBuffers) size(port int) (size int, open bool) {
	buffers.cond.L.Lock()
	defer buffers.cond.L.Unlock()

	size, open = buffers.sizes[port]
	return
}

func (buffers *remoteBuffers) reset(port int) {
	buffers.cond.L.Lock()
	defer buffers.cond.L.Unlock()

	delete(buffers.sizes, port)
	buffers.cond.Broadcast()
}

func (buffers *remoteBuffers) grow(port int, size int) {
	buffers.cond.L.Lock()
	defer buffers.cond.L.Unlock()

	buffers.sizes[port] = buffers.sizes[port] + size
	buffers.cond.Broadcast()
}

// wait waits for port's buffer to be at least size bytes and returns nil. If the link closes while wait is waiting,
// it will return the error with which the link was closed.
func (buffers *remoteBuffers) wait(port int, size int) error {
	buffers.cond.L.Lock()
	defer buffers.cond.L.Unlock()

	for {
		if buffers.link.err != nil {
			return buffers.link.err
		}
		if s, ok := buffers.sizes[port]; ok && s >= size {
			return nil
		}
		buffers.cond.Wait()
	}
}
