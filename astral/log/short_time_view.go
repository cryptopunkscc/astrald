package log

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type ShortTimeView struct {
	*astral.Time
}

func (v ShortTimeView) Render() string {
	return v.Time.Time().Format("15:04:05.000")
}

func init() {
	Set(func(o *astral.Time) astral.Object {
		return &ShortTimeView{o}
	})
}
