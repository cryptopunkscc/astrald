package linkinfo

import (
	"github.com/cryptopunkscc/astrald/infra"
	"time"
)

type Info struct {
	ReflectAddr infra.Addr
	Addrs       []AddrSpec
}

type AddrSpec struct {
	Addr      infra.Addr
	ExpiresAt time.Time
	Public    bool
}

type jsonInfo struct {
	ReflectAddr jsonAddr       `json:"reflect_addr"`
	AddrList    []jsonAddrSpec `json:"addr_list"`
}

type jsonAddr struct {
	Network string `json:"network"`
	Address []byte `json:"address"`
}

type jsonAddrSpec struct {
	Network   string `json:"network"`
	Address   []byte `json:"address"`
	Public    bool   `json:"public"`
	ExpiresAt int
}
