package channel

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
)

// TestDecode_UnknownType_WrapsBothSentinels pins the top-level-desync contract:
// astral.Decode must return an error that carries both ErrStreamCorrupted ("don't
// read more") and ErrBlueprintNotFound ("couldn't resolve this type"). Existing
// errors.Is(err, ErrBlueprintNotFound) call sites must keep working.
func TestDecode_UnknownType_WrapsBothSentinels(t *testing.T) {
	var buf bytes.Buffer
	_, err := astral.ShortTypeEncoder(&buf, "unregistered.x.for.test")
	if err != nil {
		t.Fatalf("encode tag: %v", err)
	}

	_, _, err = astral.Decode(&buf)
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !errors.Is(err, astral.ErrStreamCorrupted) {
		t.Errorf("want errors.Is(ErrStreamCorrupted), got %v", err)
	}
	if !errors.Is(err, astral.ErrBlueprintNotFound) {
		t.Errorf("want errors.Is(ErrBlueprintNotFound), got %v", err)
	}
}

// TestCanonicalReceiver_LatchesAfterUnknownType: canonical wire has no per-object
// framing, so after Stamp+tag are consumed for an unknown type the underlying reader
// is past the boundary and further reads would be garbage. The receiver latches the
// error and refuses to touch the reader on subsequent calls.
func TestCanonicalReceiver_LatchesAfterUnknownType(t *testing.T) {
	var buf bytes.Buffer
	if _, err := (astral.Stamp{}).WriteTo(&buf); err != nil {
		t.Fatalf("stamp: %v", err)
	}
	if _, err := astral.ObjectType("unregistered.x.canonical").WriteTo(&buf); err != nil {
		t.Fatalf("type: %v", err)
	}
	// some trailing garbage that must NOT be consumed after the latch fires
	buf.WriteString("trailing-garbage")

	cr := &countingReader{r: &buf}
	rcv := NewCanonicalReceiver(cr)

	_, first := rcv.Receive()
	if !errors.Is(first, astral.ErrStreamCorrupted) {
		t.Fatalf("first: want ErrStreamCorrupted, got %v", first)
	}
	readAfterFirst := cr.bytesRead

	_, second := rcv.Receive()
	if !errors.Is(second, astral.ErrStreamCorrupted) {
		t.Fatalf("second: want latched ErrStreamCorrupted, got %v", second)
	}
	if cr.bytesRead != readAfterFirst {
		t.Fatalf("latch leaked: reader consumed %d bytes between calls",
			cr.bytesRead-readAfterFirst)
	}
}

// TestJSONReceiver_LatchesAfterUnknownType: although json.Decoder is document-framed
// and could in principle continue, the policy is fail-fast — the first error latches
// and subsequent Receive() returns the same error.
func TestJSONReceiver_LatchesAfterUnknownType(t *testing.T) {
	stream := `{"Type":"unregistered.x.json","Object":null}` + "\n" +
		`{"Type":"astral.string16","Object":"\"would-have-decoded\""}` + "\n"

	cr := &countingReader{r: bytes.NewBufferString(stream)}
	rcv := NewJSONReceiver(cr)

	_, first := rcv.Receive()
	if !errors.Is(first, astral.ErrStreamCorrupted) {
		t.Fatalf("first: want ErrStreamCorrupted, got %v", first)
	}
	readAfterFirst := cr.bytesRead

	_, second := rcv.Receive()
	if !errors.Is(second, astral.ErrStreamCorrupted) {
		t.Fatalf("second: want latched ErrStreamCorrupted, got %v", second)
	}
	if cr.bytesRead != readAfterFirst {
		t.Fatalf("latch leaked: reader consumed %d bytes between calls",
			cr.bytesRead-readAfterFirst)
	}
}

// TestTextReceiver_LatchesAfterUnknownType: lines are self-delimited, but the policy
// is fail-fast — the first error latches.
func TestTextReceiver_LatchesAfterUnknownType(t *testing.T) {
	stream := "#[unregistered.x.text]\n" +
		"#[astral.string16] would-have-decoded\n"

	cr := &countingReader{r: bytes.NewBufferString(stream)}
	rcv := NewTextReceiver(cr)

	_, first := rcv.Receive()
	if !errors.Is(first, astral.ErrStreamCorrupted) {
		t.Fatalf("first: want ErrStreamCorrupted, got %v", first)
	}
	readAfterFirst := cr.bytesRead

	_, second := rcv.Receive()
	if !errors.Is(second, astral.ErrStreamCorrupted) {
		t.Fatalf("second: want latched ErrStreamCorrupted, got %v", second)
	}
	if cr.bytesRead != readAfterFirst {
		t.Fatalf("latch leaked: reader consumed %d bytes between calls",
			cr.bytesRead-readAfterFirst)
	}
}

// TestBinaryReceiver_AllowUnparsed_TopLevel_StillWorks: top-level unknown types are
// substituted by *UnparsedObject when AllowUnparsed=true, because Bytes32 framing
// keeps the payload local to the receiver.
func TestBinaryReceiver_AllowUnparsed_TopLevel_StillWorks(t *testing.T) {
	var buf bytes.Buffer
	if _, err := astral.ObjectType("unregistered.x.binary").WriteTo(&buf); err != nil {
		t.Fatalf("type: %v", err)
	}
	payload := astral.Bytes32([]byte{0xDE, 0xAD, 0xBE, 0xEF})
	if _, err := payload.WriteTo(&buf); err != nil {
		t.Fatalf("payload: %v", err)
	}

	br := NewBinaryReceiver(&buf)
	br.AllowUnparsed = true

	obj, err := br.Receive()
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	up, ok := obj.(*astral.UnparsedObject)
	if !ok {
		t.Fatalf("want *UnparsedObject, got %T", obj)
	}
	if up.Type != "unregistered.x.binary" {
		t.Errorf("Type: want %q, got %q", "unregistered.x.binary", up.Type)
	}
	if !bytes.Equal(up.Payload, []byte{0xDE, 0xAD, 0xBE, 0xEF}) {
		t.Errorf("Payload: want %x, got %x", []byte{0xDE, 0xAD, 0xBE, 0xEF}, up.Payload)
	}
}

// TestBinaryReceiver_AllowUnparsed_NestedUnknown_StillWorks: a registered Blueprint
// with an ObjectSpec field receives an unknown nested type tag. Decoding the inner
// envelope returns an error wrapping ErrStreamCorrupted + ErrBlueprintNotFound; the
// receiver catches errors.Is(ErrBlueprintNotFound) and returns the whole outer
// envelope as *UnparsedObject. The sentinel wrap does not break the existing
// recovery branch.
func TestBinaryReceiver_AllowUnparsed_NestedUnknown_StillWorks(t *testing.T) {
	bp := astral.NewBlueprint("test.stream_corrupted.outer",
		astral.Field{Name: "Inner", Spec: &astral.ObjectSpec{}},
	)
	if _, err := astral.Register(bp); err != nil {
		t.Fatalf("register: %v", err)
	}

	// inner = [String8 "unknown.x.nested"][no payload bytes]
	var inner bytes.Buffer
	if _, err := astral.String8("unknown.x.nested").WriteTo(&inner); err != nil {
		t.Fatalf("inner tag: %v", err)
	}

	// envelope = [outer ObjectType][Bytes32 of inner.Bytes()]
	var envelope bytes.Buffer
	if _, err := astral.ObjectType("test.stream_corrupted.outer").WriteTo(&envelope); err != nil {
		t.Fatalf("envelope tag: %v", err)
	}
	if _, err := astral.Bytes32(inner.Bytes()).WriteTo(&envelope); err != nil {
		t.Fatalf("envelope payload: %v", err)
	}

	br := NewBinaryReceiver(&envelope)
	br.AllowUnparsed = true

	obj, err := br.Receive()
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	up, ok := obj.(*astral.UnparsedObject)
	if !ok {
		t.Fatalf("want *UnparsedObject after nested unknown, got %T (err=%v)", obj, err)
	}
	if up.Type != "test.stream_corrupted.outer" {
		t.Errorf("Type: want outer envelope, got %q", up.Type)
	}
	if !bytes.Equal(up.Payload, inner.Bytes()) {
		t.Errorf("Payload: want full inner envelope, got %x", up.Payload)
	}
}

// countingReader observes how many bytes a receiver pulls from the underlying transport.
// Used by latch tests to assert that a latched receiver does not touch the reader on
// subsequent Receive() calls.
type countingReader struct {
	r         io.Reader
	bytesRead int
}

func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.bytesRead += n
	return n, err
}
