package proto

import "github.com/cryptopunkscc/astrald/auth/id"

type ServiceEntry struct {
	Identity id.Identity `cslq:"v"`
	Name     string      `cslq:"[c]c"`
	Type     string      `cslq:"[c]c"`
	Extra    []byte      `cslq:"[s]c"`
}

type MsgRegister struct {
	Service string `cslq:"[c]c"`
}
