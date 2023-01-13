package nodeinfo

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/jxskiss/base62"
	"strings"
)

type NodeInfo struct {
	Identity  id.Identity
	Alias     string
	Addresses []infra.Addr
}

func New(identity id.Identity) *NodeInfo {
	return &NodeInfo{
		Identity:  identity,
		Addresses: make([]infra.Addr, 0),
	}
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
