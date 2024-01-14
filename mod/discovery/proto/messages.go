package proto

import "github.com/cryptopunkscc/astrald/auth/id"

type Info struct {
	Data     []Data    `cslq:"[c]v"`
	Services []Service `cslq:"[c]v"`
}

type Data struct {
	Bytes []byte `cslq:"[s]c"`
}

type Service struct {
	Identity id.Identity `cslq:"v"`
	Name     string      `cslq:"[c]c"`
	Type     string      `cslq:"[c]c"`
	Extra    []byte      `cslq:"[s]c"`
}
