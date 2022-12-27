package mux

import (
	"errors"
	"io"
	"sync"
	"testing"
)

func TestDemux(t *testing.T) {
	var msg = []byte("test")
	var setup sync.WaitGroup
	var workers sync.WaitGroup

	r, w := io.Pipe()
	mux := NewMux(w)
	demux := NewStreamDemux(r)

	setup.Add(3)
	workers.Add(4)
	for i := 0; i < 3; i++ {
		go func(i int) {
			defer workers.Done()
			var input *InputStream
			var err error

			if i == 0 {
				input, err = demux.ControlStream()
			} else {
				input, err = demux.AllocStream()
			}
			if err != nil {
				t.Fatal(err)
			}

			var buf = make([]byte, len(msg))

			setup.Done()

			n, err := input.Read(buf)
			if err != nil {
				t.Fatal(err)
			}
			if n != len(msg) {
				t.Fatal("default stream invalid data length")
			}

			n, err = input.Read(buf)
			if !errors.Is(err, io.EOF) {
				t.Fatal("unexpected error:", err)
			}
			if n > 0 {
				t.Fatal("unexpected data")
			}
		}(i)
	}

	go func() {
		defer workers.Done()
		setup.Wait() // wait for the demux to set up
		var err error

		if err = mux.Write(0, msg); err != nil {
			t.Fatal(err)
		}
		if err = mux.Write(1, msg); err != nil {
			t.Fatal(err)
		}
		if err = mux.Write(2, msg); err != nil {
			t.Fatal(err)
		}
		mux.Close()
	}()

	workers.Wait()
}
