package log

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type EntryView struct {
	*Entry
}

func (v EntryView) Render() string {
	var line = Format("[%v] (%v) %v ",
		v.Origin,
		&v.Level,
		&v.Time,
	)

	line = append(line, v.Objects...)

	return DefaultViewer.Render(line...)
}

func init() {
	DefaultViewer.Set(Entry{}.ObjectType(), func(o astral.Object) astral.Object {
		return EntryView{o.(*Entry)}
	})
}
