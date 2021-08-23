package uid

import (
	"github.com/cryptopunkscc/astrald/api"
	"reflect"
	"testing"
)

func TestPackUnpack(t *testing.T) {
	// given
	expected := Card{
		Id: api.Identity("id"),
		Alias: "alias",
		Endpoints: []Endpoint{
			{"net1", "addr1"},
			{"net2", "addr2"},
		},
	}

	// when
	result := Unpack(Pack(expected))

	// then
	if !reflect.DeepEqual(expected, result) {
		t.Error("not equal", expected, result)
	}
}
