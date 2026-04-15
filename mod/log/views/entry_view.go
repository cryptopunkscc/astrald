package views

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type EntryView struct {
	*log.Entry
}

var HideOrigin *astral.Identity

func (v EntryView) Render() string {
	level := fmt.Sprintf("(%v)", v.Level)

	var line = log.Format("%v %v ",
		String(level, &styles.DarkGrayText),
		NewTimeView(&v.Time),
	)

	if HideOrigin == nil || !v.Origin.IsEqual(HideOrigin) {
		line = append(log.Format("[%v] ", v.Origin), line...)
	}

	line = append(line, v.Objects...)

	return log.DefaultViewer.Render(line...)
}

func UseEntryView() {
	log.DefaultViewer.Set(log.Entry{}.ObjectType(), func(o astral.Object) astral.Object {
		return EntryView{o.(*log.Entry)}
	})
}
