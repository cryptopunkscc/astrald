package proto

type ServiceEntry struct {
	Name  string `cslq:"[c]c"`
	Type  string `cslq:"[c]c"`
	Extra []byte `cslq:"[s]c"`
}

type MsgRegister struct {
	Service string `cslq:"[c]c"`
}
