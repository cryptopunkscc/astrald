package channel

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
)

// TestTextChannel_RuntimeObject_RoundTrip pins the base64-via-Text fallback for
// RuntimeObject: TextSender sees no TextMarshaler and writes "#[type]:<base64>\n";
// TextReceiver parses the ":" marker, base64-decodes the payload, and runs
// RuntimeObject.ReadFrom over the binary bytes. If someone later adds MarshalText
// to RuntimeObject the wire shape will shift from base64 to type-specific and this
// test will catch it.
func TestTextChannel_RuntimeObject_RoundTrip(t *testing.T) {
	bp := astral.NewBlueprint("test.channel.text.msg",
		astral.Field{Name: "Text", Spec: &astral.PrimitiveSpec{PrimitiveType: "string16"}},
		astral.Field{Name: "Count", Spec: &astral.PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	_, err := astral.RegisterBlueprint(bp)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	src, err := bp.GetRuntimeObject()
	if err != nil {
		t.Fatal(err)
	}
	_ = src.Set("Text", "hello")
	_ = src.Set("Count", uint32(3))

	var buf bytes.Buffer
	sender := NewTextSender(&buf)
	if err := sender.Send(src); err != nil {
		t.Fatalf("send: %v", err)
	}

	// pin the wire shape: "#[type]:<base64>\n" — base64 lane (":"), not text (" ").
	line := buf.String()
	prefix := "#[test.channel.text.msg]:"
	if !strings.HasPrefix(line, prefix) {
		t.Fatalf("wire shape: want prefix %q, got %q", prefix, line)
	}
	if !strings.HasSuffix(line, "\n") {
		t.Fatalf("wire shape: missing trailing newline in %q", line)
	}

	recv := NewTextReceiver(&buf)
	got, err := recv.Receive()
	if err != nil {
		t.Fatalf("receive: %v", err)
	}

	ro, ok := got.(*astral.RuntimeObject)
	if !ok {
		t.Fatalf("want *astral.RuntimeObject, got %T", got)
	}
	if s, _ := ro.Get("Text").(*astral.String16); s == nil || *s != "hello" {
		t.Fatalf("Text: want %q, got %#v", "hello", ro.Get("Text"))
	}
	if u, _ := ro.Get("Count").(*astral.Uint32); u == nil || *u != 3 {
		t.Fatalf("Count: want 3, got %#v", ro.Get("Count"))
	}
}
