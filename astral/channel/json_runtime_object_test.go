package channel

import (
	"bytes"
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
)

// TestJSONChannel_RuntimeObject_RoundTrip exercises the silent-loss scenario directly:
// a runtime-registered Blueprint, populated, sent through JSONSender, received through
// JSONReceiver. Before RuntimeObject grew MarshalJSON/UnmarshalJSON the receiver got a
// spec-zero object with no error; this test pins that field values now survive the trip.
func TestJSONChannel_RuntimeObject_RoundTrip(t *testing.T) {
	bp := astral.NewBlueprint("test.channel.json.msg",
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
	sender := NewJSONSender(&buf)
	if err := sender.Send(src); err != nil {
		t.Fatalf("send: %v", err)
	}

	recv := NewJSONReceiver(&buf)
	got, err := recv.Receive()
	if err != nil {
		t.Fatalf("receive: %v", err)
	}

	ro, ok := got.(*astral.RuntimeObject)
	if !ok {
		t.Fatalf("want *astral.RuntimeObject, got %T", got)
	}
	if s, _ := ro.Get("Text").(*astral.String16); s == nil || *s != "hello" {
		t.Fatalf("Text: want \"hello\", got %#v", ro.Get("Text"))
	}
	if u, _ := ro.Get("Count").(*astral.Uint32); u == nil || *u != 3 {
		t.Fatalf("Count: want 3, got %#v", ro.Get("Count"))
	}
}
