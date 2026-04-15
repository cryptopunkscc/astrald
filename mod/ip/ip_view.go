package ip

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type IPView struct {
	*IP
}

func (v IPView) Render() string {
	return styles.BrightBlueText.Render(v.IP.String())
}

func init() {
	log.DefaultViewer.Set(IP(nil).ObjectType(), func(object astral.Object) astral.Object {
		return &IPView{object.(*IP)}
	})
}
