package log

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type EntryView struct {
	*log.Entry
}

func (v EntryView) Render() string {
	level := fmt.Sprintf("(%v)", v.Level)

	var line = log.Format("[%v] %v %v ",
		v.Origin,
		String(level, &DarkGrayText),
		NewTimeViewWithStyle(&v.Time, "", GrayText),
	)

	line = append(line, v.Objects...)

	return log.DefaultViewer.Render(line...)
}

func UseEntryView() {
	log.DefaultViewer.Set(log.Entry{}.ObjectType(), func(o astral.Object) astral.Object {
		return EntryView{o.(*log.Entry)}
	})
}
