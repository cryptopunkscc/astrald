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

	demux := NewFrameDemux(bytes.NewReader(buf.Bytes()))

	frame, err := demux.ReadFrame()
	if (frame.StreamID != 1) || (bytes.Compare(frame1, frame.Data) != 0) || (err != nil) {
		t.Error("case 1 failed")
	}

	frame, err = demux.ReadFrame()
	if (frame.StreamID != 2) || (bytes.Compare(frame2, frame.Data) != 0) || (err != nil) {
		t.Error("case 2 failed")
	}

	frame, err = demux.ReadFrame()
	if (frame.StreamID != 3) || (bytes.Compare(frame3, frame.Data) != 0) || (err != nil) {
		t.Error("case 3 failed")
	}
}
