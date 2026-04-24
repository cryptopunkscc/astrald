package views

import (
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type EntryView struct {
	*log.Entry
}

func (v EntryView) Render() string {
	out := fmt.Sprintf("[%v] (%d) %v ",
		v.Origin,
		v.Level,
		&v.Time,
	)

	for _, object := range v.Objects {
		out += fmt.Sprint(object)
	}

	return out
}

func init() {
	fmt.SetView(func(o *log.Entry) fmt.View {
		return EntryView{Entry: o}
	})
}
