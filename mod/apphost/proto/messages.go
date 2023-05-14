package proto

import (
	"github.com/cryptopunkscc/astrald/auth/id"
)

const (
	CmdRegister = "register"
	CmdQuery    = "query"
	CmdResolve  = "resolve"
	CmdNodeInfo = "nodeInfo"
)

type Command struct {
	Cmd string `cslq:"[c]c"`
}

type AuthParams struct {
	Token string `cslq:"[c]c"`
}

type QueryParams struct {
	Identity id.Identity `cslq:"v"`
	Query    string      `cslq:"[c]c"`
}

type RegisterParams struct {
	Service string `cslq:"[c]c"`
	Target  string `cslq:"[c]c"`
}

type NodeInfoParams struct {
	Identity id.Identity `cslq:"v"`
}

type NodeInfoData struct {
	Identity id.Identity `cslq:"v"`
	Name     string      `cslq:"[c]c"`
}

type InQueryParams struct {
	Identity id.Identity `cslq:"v"`
	Query    string      `cslq:"[c]c"`
}

type ResolveParams struct {
	Name string `cslq:"[c]c"`
}

type ResolveData struct {
	Identity id.Identity `cslq:"v"`
}
