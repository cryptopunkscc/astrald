package message

import (
	"testing"
)

func TestPackUnpack(t *testing.T) {
	given := NewMessage("sender", "recipient", "text")

	// when
	packed := given.Pack()
	unpacked, err := Unpack(packed)
	if err != nil {
		t.Error(err)
	}
	result := unpacked

	// then
	if given.Sender() != result.Sender() {
		t.Error("not equal", given.Sender(), result.Sender())
	}
	if given.Recipient() != result.Recipient() {
		t.Error("not equal", given.Recipient(), result.Recipient())
	}
	if given.Text() != result.Text() {
		t.Error("not equal", given.Text(), result.Text())
	}
}
