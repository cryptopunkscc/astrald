package streams

import (
	"io"
	"sync"
)

// Dispenser is an io.WriteCloser that writes a limited number of bytes to the output.
type Dispenser struct {
	limit     int
	cond      *sync.Cond
	output    io.WriteCloser
	unlimited bool
	closed    bool
}

// NewDispenser makes a new Dispenser that writes to the provided output.
func NewDispenser(output io.WriteCloser) *Dispenser {
	return &Dispenser{
		output: output,
		cond:   sync.NewCond(&sync.Mutex{}),
	}
}

// Write writes at most Limit() bytes to the output. If len(p) is larger than Limit(), Write will block
// until the limit increases or the writer closes. While Write() is waiting for the limit to increase it's ok
// to call SetOutput().
func (w *Dispenser) Write(p []byte) (n int, err error) {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	if w.closed {
		return n, io.ErrClosedPipe
	}
	if w.unlimited {
		return w.output.Write(p)
	}

	for len(p) > 0 {
		if w.unlimited {
			nn, err := w.output.Write(p)
			return n + nn, err
		}
		if w.closed {
			return n, io.ErrClosedPipe
		}
		if w.limit == 0 {
			w.cond.Wait()
			continue
		}

		l := min(len(p), w.limit)
		nn, err := w.output.Write(p[:l])
		n += nn
		if err != nil {
			return n, err
		}
		p = p[nn:]
		w.limit -= nn
		w.cond.Broadcast()
	}

	return
}

// SetOutput sets the target WriteCloser.
func (w *Dispenser) SetOutput(output io.WriteCloser) {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	w.output = output
}

// SetUnlimited if set to true, Write will not obey the limit and simply pass through any Write calls directly
// to the output.
func (w *Dispenser) SetUnlimited(v bool) error {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	w.unlimited = v

	w.cond.Broadcast()
	return nil
}

// Limit returns the reamining write limit.
func (w *Dispenser) Limit() int {
	return w.limit
}

// Increase increases the write limit by i bytes.
func (w *Dispenser) Increase(i int) {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	w.limit += i
	w.cond.Broadcast()
}

// Flush blocks until the limit reaches 0 or the writer closes.
func (w *Dispenser) Flush() {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	for {
		if w.limit == 0 || w.closed {
			return
		}
		w.cond.Wait()
	}
}

// Close closes the writer and aborts any blocking Writes.
func (w *Dispenser) Close() error {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	w.closed = true
	w.cond.Broadcast()
	return w.output.Close()
}
