package nodeinfo

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"testing"
)

func TestMarshalling(t *testing.T) {
	var testID, _ = id.GenerateIdentity()
	var nodeInfo = New(testID)

	nodeInfo.Alias = "tester"
	nodeInfo.Endpoints = append(nodeInfo.Endpoints, net.NewGenericEndpoint("test", []byte{1, 2, 3, 4}))
	nodeInfo.Endpoints = append(nodeInfo.Endpoints, net.NewGenericEndpoint("tset", []byte{4, 3, 2, 1}))

	var s = nodeInfo.String()

	var read, err = Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	if read.Alias != nodeInfo.Alias {
		t.Fatal("alias mismatch")
	}
	if len(read.Endpoints) != len(nodeInfo.Endpoints) {
		t.Fatal("address count mismatch")
	}
	for i := range read.Endpoints {
		if !net.EndpointEqual(read.Endpoints[i], nodeInfo.Endpoints[i]) {
			t.Fatal("address")
		}
	}
}
