package mux

import (
	"errors"
	"io"
	"testing"
)

func TestClosingWriter(t *testing.T) {
	var message = "test"
	var input = newInputStream(nil, 0)
	var buf = make([]byte, 32)

	go func() {
		input.write([]byte(message))
		input.closeWriter(nil)
	}()

	n, err := input.Read(buf)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if n != len(message) {
		t.Fatal("read invalid message length")
	}

	_, err = input.Read(buf)
	if !errors.Is(err, io.EOF) {
		t.Fatal("expected EOF, got", err)
	}
}
