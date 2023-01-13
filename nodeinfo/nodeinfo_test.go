package nodeinfo

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"testing"
)

func TestMarshalling(t *testing.T) {
	var testID, _ = id.GenerateIdentity()
	var nodeInfo = New(testID)

	nodeInfo.Alias = "tester"
	nodeInfo.Addresses = append(nodeInfo.Addresses, infra.NewGenericAddr("test", []byte{1, 2, 3, 4}))
	nodeInfo.Addresses = append(nodeInfo.Addresses, infra.NewGenericAddr("tset", []byte{4, 3, 2, 1}))

	var s = nodeInfo.String()

	var read, err = Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	if read.Alias != nodeInfo.Alias {
		t.Fatal("alias mismatch")
	}
	if len(read.Addresses) != len(nodeInfo.Addresses) {
		t.Fatal("address count mismatch")
	}
	for i := range read.Addresses {
		if !infra.AddrEqual(read.Addresses[i], nodeInfo.Addresses[i]) {
			t.Fatal("address")
		}
	}
}
