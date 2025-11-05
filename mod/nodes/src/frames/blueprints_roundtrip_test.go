package frames

import (
	"reflect"
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
)

func TestFrameBlueprintsRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		obj  astral.Object
	}{
		{"ping", &Ping{Nonce: astral.NewNonce(), Pong: false}},
		{"pong", &Ping{Nonce: astral.NewNonce(), Pong: true}},
		{"query", &Query{Nonce: astral.NewNonce(), Buffer: 12345, Query: "hello"}},
		{"response", &Response{Nonce: astral.NewNonce(), ErrCode: 1, Buffer: 10}},
		{"read", &Read{Nonce: astral.NewNonce(), Len: 4096}},
		{"data", &Data{Nonce: astral.NewNonce(), Payload: []byte("payload")}},
		{"migrate", &Migrate{Nonce: astral.NewNonce()}},
		{"reset", &Reset{Nonce: astral.NewNonce()}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := FrameBlueprints.Pack(tc.obj)
			if err != nil {
				t.Fatalf("pack failed: %v", err)
			}

			obj, err := FrameBlueprints.Unpack(data)
			if err != nil {
				t.Fatalf("unpack failed: %v", err)
			}

			if reflect.TypeOf(obj) != reflect.TypeOf(tc.obj) {
				t.Fatalf("type mismatch: got %T want %T", obj, tc.obj)
			}

			if !reflect.DeepEqual(obj, tc.obj) {
				t.Fatalf("value mismatch: got %#v want %#v", obj, tc.obj)
			}
		})
	}
}
