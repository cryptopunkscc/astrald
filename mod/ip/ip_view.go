package ip

import (
	"github.com/cryptopunkscc/astrald/astral"
	log2 "github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log"
)

type IPView struct {
	*IP
}

func (v IPView) Render() string {
	return log.BrightBlueText.Render(v.IP.String())
}

func init() {
	log2.DefaultViewer.Set(IP(nil).ObjectType(), func(object astral.Object) astral.Object {
		return &IPView{object.(*IP)}
	})
}
