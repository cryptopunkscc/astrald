package nodeinfo

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/jxskiss/base62"
	"strings"
)

type NodeInfo struct {
	Identity  id.Identity
	Alias     string
	Endpoints []net.Endpoint
}

func New(identity id.Identity) *NodeInfo {
	info := &NodeInfo{
		Identity:  identity,
		Alias:     "",
		Endpoints: make([]net.Endpoint, 0),
	}

	return info
}

func FromNode(node node.Node) *NodeInfo {
	info := &NodeInfo{
		Identity:  node.Identity().Public(),
		Alias:     node.Alias(),
		Endpoints: node.Infra().Endpoints(),
	}

	return info
}

func Parse(s string) (*NodeInfo, error) {
	trimmed := strings.TrimPrefix(s, infoPrefix)
	data, err := base62.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}

	nodeInfo := &NodeInfo{}
	if err := cslq.Decode(bytes.NewReader(data), "v", nodeInfo); err != nil {
		return nil, err
	}

	return nodeInfo, nil
}

func (info *NodeInfo) String() string {
	var buf = &bytes.Buffer{}
	if err := cslq.Encode(buf, "v", info); err != nil {
		return "error"
	}
	return infoPrefix + base62.EncodeToString(buf.Bytes())
}
