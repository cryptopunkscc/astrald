package story

import (
	fid2 "github.com/cryptopunkscc/astrald/components/fid"
	"reflect"
	"testing"
)

func TestPackUnpack(t *testing.T) {
	given := Story{
		timestamp: uint64(1),
		typ:       []byte{2},
		author:    []byte{3},
		refs: []fid2.ID{
			fid2.ResolveBytes([]byte{4}),
			fid2.ResolveBytes([]byte{5}),
		},
		data: []byte{6},
	}

	// when
	packed := given.Pack()
	unpacked, err := UnpackBytes(packed)
	if err != nil {
		t.Error(err)
	}
	result := &unpacked

	// then
	if reflect.DeepEqual(result, given) {
		t.Error("not equal", given, result)
	}
}
