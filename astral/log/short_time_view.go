package log

import "github.com/cryptopunkscc/astrald/astral"

type ShortTimeView struct {
	*astral.Time
}

func (v ShortTimeView) Render() string {
	return v.Time.Time().Format("15:04:05.000")
}

func init() {
	DefaultViewer.Set(astral.Time{}.ObjectType(), func(o astral.Object) astral.Object {
		return &ShortTimeView{o.(*astral.Time)}
	})
}
