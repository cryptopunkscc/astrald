package rpc

import (
	"github.com/cryptopunkscc/astrald/data"
)

type MsgRegisterSource struct {
	Service string `cslq:"[c]c"`
}

type MsgRead struct {
	DataID data.ID `cslq:"v"`
	Start  int64   `cslq:"q"`
	Len    int64   `cslq:"q"`
}
