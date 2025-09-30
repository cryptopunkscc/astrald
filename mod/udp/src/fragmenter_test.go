package udp

import (
	"testing"
)

func TestBasicFragmenter_SingleFragment(t *testing.T) {
	mss := 100
	frag := NewBasicFragmenter(mss)
	data := make([]byte, 80)
	for i := range data {
		data[i] = byte(i)
	}
	buf := &ByteStreamBuffer{data: data}
	packet, packetLen, ok := frag.MakeNew(0, mss, buf)
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if packetLen != 80 {
		t.Errorf("expected packetLen=80, got %d", packetLen)
	}
	if packet == nil {
		t.Fatalf("expected non-nil packet")
	}
	if packet.Seq != 0 {
		t.Errorf("expected Seq=0, got %d", packet.Seq)
	}
	if packet.Len != 80 {
		t.Errorf("expected Len=80, got %d", packet.Len)
	}
	for i := range packet.Payload {
		if packet.Payload[i] != byte(i) {
			t.Errorf("payload mismatch at %d: got %d, want %d", i, packet.Payload[i], byte(i))
		}
	}
}

func TestBasicFragmenter_MultipleFragments(t *testing.T) {
	mss := 50
	frag := NewBasicFragmenter(mss)
	data := make([]byte, 120)
	for i := range data {
		data[i] = byte(i)
	}
	buf := &ByteStreamBuffer{data: data}

	nextSeq := uint32(0)
	total := 0
	fragments := 0
	for buf.Len() > 0 {
		packet, packetLength, ok := frag.MakeNew(nextSeq, mss, buf)
		if !ok {
			t.Fatalf("expected ok=true, got false")
		}
		if packet == nil {
			t.Fatalf("expected non-nil packet")
		}
		if int(packet.Len) > mss {
			t.Errorf("fragment too large: got %d, want <= %d", packet.Len, mss)
		}
		for i := 0; i < int(packet.Len); i++ {
			if packet.Payload[i] != byte(int(nextSeq)+i) {
				t.Errorf("payload mismatch at %d: got %d, want %d", int(nextSeq)+i, packet.Payload[i], byte(int(nextSeq)+i))
			}
		}
		buf.Advance(packetLength)
		nextSeq += uint32(packetLength)
		total += packetLength
		fragments++
	}
	if total != 120 {
		t.Errorf("expected total=120, got %d", total)
	}
	if fragments != 3 {
		t.Errorf("expected fragments=3, got %d", fragments)
	}
}

func TestBasicFragmenter_ZeroLen(t *testing.T) {
	mss := 50
	frag := NewBasicFragmenter(mss)
	buf := &ByteStreamBuffer{data: nil}
	packet, packetLength, ok := frag.MakeNew(0, mss, buf)
	if ok {
		t.Errorf("expected ok=false for zero-len buffer")
	}
	if packet != nil {
		t.Errorf("expected nil packet for zero-len buffer")
	}
	if packetLength != 0 {
		t.Errorf("expected packetLength=0 for zero-len buffer, got %d", packetLength)
	}
}

func TestBasicFragmenter_AllowedLessThanMSS(t *testing.T) {
	mss := 100
	allowed := 40
	frag := NewBasicFragmenter(mss)
	data := make([]byte, 80)
	for i := range data {
		data[i] = byte(i)
	}
	buf := &ByteStreamBuffer{data: data}
	packet, packetLength, ok := frag.MakeNew(0, allowed, buf)
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if packetLength != allowed {
		t.Errorf("expected packetLength=%d, got %d", allowed, packetLength)
	}
	if packet.Len != uint16(allowed) {
		t.Errorf("expected Len=%d, got %d", allowed, packet.Len)
	}
}

func TestBasicFragmenter_AllowedLessThanBuffer(t *testing.T) {
	mss := 100
	allowed := 60
	frag := NewBasicFragmenter(mss)
	data := make([]byte, 80)
	for i := range data {
		data[i] = byte(i)
	}
	buf := &ByteStreamBuffer{data: data}
	packet, packetLength, ok := frag.MakeNew(0, allowed, buf)
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if packetLength != allowed {
		t.Errorf("expected packetLength=%d, got %d", allowed, packetLength)
	}
	if packet.Len != uint16(allowed) {
		t.Errorf("expected Len=%d, got %d", allowed, packet.Len)
	}
}

func TestBasicFragmenter_BufferSmallerThanAllowedAndMSS(t *testing.T) {
	mss := 100
	allowed := 80
	frag := NewBasicFragmenter(mss)
	data := make([]byte, 50)
	for i := range data {
		data[i] = byte(i)
	}
	buf := &ByteStreamBuffer{data: data}
	packet, packetLength, ok := frag.MakeNew(0, allowed, buf)
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if packetLength != 50 {
		t.Errorf("expected packetLength=50, got %d", packetLength)
	}
	if packet.Len != 50 {
		t.Errorf("expected Len=50, got %d", packet.Len)
	}
}

func TestBasicFragmenter_NegativeOrZeroAllowed(t *testing.T) {
	mss := 100
	frag := NewBasicFragmenter(mss)
	data := make([]byte, 50)
	buf := &ByteStreamBuffer{data: data}
	packet, packetLength, ok := frag.MakeNew(0, 0, buf)
	if ok || packet != nil || packetLength != 0 {
		t.Errorf("expected no packet for allowed=0")
	}
	packet, packetLength, ok = frag.MakeNew(0, -10, buf)
	if ok || packet != nil || packetLength != 0 {
		t.Errorf("expected no packet for allowed<0")
	}
}

func TestBasicFragmenter_ZeroMSS(t *testing.T) {
	mss := 0
	frag := NewBasicFragmenter(mss)
	data := make([]byte, 50)
	buf := &ByteStreamBuffer{data: data}
	packet, packetLength, ok := frag.MakeNew(0, 100, buf)
	if ok || packet != nil || packetLength != 0 {
		t.Errorf("expected no packet for MSS=0")
	}
}

func TestBasicFragmenter_FlagsSet(t *testing.T) {
	mss := 100
	frag := NewBasicFragmenter(mss)
	data := make([]byte, 50)
	buf := &ByteStreamBuffer{data: data}
	packet, _, ok := frag.MakeNew(0, 100, buf)
	if !ok || packet == nil {
		t.Fatalf("expected valid packet")
	}
	if packet.Flags&FlagACK == 0 {
		t.Errorf("expected ACK flag set in data packet")
	}
}
