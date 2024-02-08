package presence

import (
	"github.com/cryptopunkscc/astrald/mod/presence"
)

type FlagOnce struct {
	mod  *Module
	flag string
}

func NewFlagOnce(mod *Module, flag string) *FlagOnce {
	return &FlagOnce{mod: mod, flag: flag}
}

func (flag *FlagOnce) OnPendingAd(ad presence.PendingAd) {
	ad.AddFlag(flag.flag)
	flag.mod.RemoveHookAdOut(flag)
}
