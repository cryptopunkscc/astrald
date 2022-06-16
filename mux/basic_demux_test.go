package mux

import (
	"bytes"
	"testing"
)

func TestReadFrame(t *testing.T) {
	var frame1 = []byte{}
	var frame2 = []byte{1, 2, 3}
	var frame3 = []byte{0, 0, 0, 0, 0, 0, 0, 0}

	buf := &bytes.Buffer{}
	mux := NewMux(buf)
	mux.Write(1, frame1)
	mux.Write(2, frame2)
	mux.Write(3, frame3)

	demux := NewBasicDemux(bytes.NewReader(buf.Bytes()))

	s, b, e := demux.ReadFrame()
	if (s != 1) || (bytes.Compare(frame1, b) != 0) || (e != nil) {
		t.Error("TestReadFrame() case 1 failed")
	}

	s, b, e = demux.ReadFrame()
	if (s != 2) || (bytes.Compare(frame2, b) != 0) || (e != nil) {
		t.Error("TestReadFrame() case 2 failed")
	}

	s, b, e = demux.ReadFrame()
	if (s != 3) || (bytes.Compare(frame3, b) != 0) || (e != nil) {
		t.Error("TestReadFrame() case 3 failed")
	}
}
