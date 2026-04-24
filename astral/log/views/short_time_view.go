package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
)

type ShortTimeView struct {
	*astral.Time
}

func (v ShortTimeView) Render() string {
	return v.Time.Time().Format("15:04:05.000")
}

func init() {
	fmt.SetView(func(o *astral.Time) fmt.View {
		return &ShortTimeView{Time: o}
	})
}
