package story

import (
	"bytes"
	fid2 "github.com/cryptopunkscc/astrald/components/fid"
	"testing"
)

func TestPackUnpack(t *testing.T) {
	given := NewStory(1, "typ", "identity", []fid2.ID{
		fid2.ResolveBytes([]byte{4}),
		fid2.ResolveBytes([]byte{5}),
	}, []byte{6})

	// when
	packed := given.Pack()
	result, err := UnpackBytes(packed)
	if err != nil {
		t.Error(err)
	}

	// then
	if given.Author() != result.Author() {
		t.Error("not equal", given, result)
	}
	if given.Timestamp() != result.Timestamp() {
		t.Error("not equal", given, result)
	}
	if given.Type() != result.Type() {
		t.Error("not equal", given, result)
	}
	if len(given.Refs()) != len(result.Refs()) {
		t.Error("not equal", given, result)
	}
	for i, id := range given.Refs() {
		if id.String() != result.Refs()[i].String() {
			t.Error("not equal", id, result.Refs()[i])
		}
	}
	givenData := bytes.NewBuffer(given.data).String()
	resultData := bytes.NewBuffer(result.data).String()

	if givenData != resultData {
		t.Error("not equal", givenData, resultData)
	}
}
