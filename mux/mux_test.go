package mux

import (
	"bytes"
	"testing"
)

func TestWrite(t *testing.T) {
	var res1 = []byte{0, 1, 0, 0}
	var res2 = []byte{0, 3, 0, 3, 3, 2, 1}
	var res3 = []byte{0, 1, 0, 1, 1, 0, 2, 0, 2, 0, 2}

	// case 1
	var buf = &bytes.Buffer{}
	mux := NewMux(buf)
	mux.Write(1, []byte{})
	if c := bytes.Compare(buf.Bytes(), res1); c != 0 {
		t.Error("TestWrite() case 1 failed")
	}

	// case 2
	buf = &bytes.Buffer{}
	mux = NewMux(buf)
	mux.Write(3, []byte{3, 2, 1})
	if c := bytes.Compare(buf.Bytes(), res2); c != 0 {
		t.Error("TestWrite() case 2 failed")
	}

	// case 3
	buf = &bytes.Buffer{}
	mux = NewMux(buf)
	mux.Write(1, []byte{1})
	mux.Write(2, []byte{0, 2})
	if c := bytes.Compare(buf.Bytes(), res3); c != 0 {
		t.Error("TestWrite() case 3 failed")
	}
}
