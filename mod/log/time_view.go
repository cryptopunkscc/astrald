package log

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type TimeView struct {
	*astral.Time
	Layout string
	Style  lipgloss.Style
}

func NewTimeView(time *astral.Time, layout string) *TimeView {
	return &TimeView{Time: time, Layout: layout, Style: DarkGrayText}
}

func NewTimeViewWithStyle(time *astral.Time, layout string, style lipgloss.Style) *TimeView {
	return &TimeView{Time: time, Layout: layout, Style: style}
}

func (v TimeView) Render() string {
	layout := v.Layout
	if layout == "" {
		layout = "15:04:05.000"
	}

	return v.Style.Render(v.Time.Time().Format(layout))
}

func init() {
	log.Set(func(o *astral.Time) astral.Object {
		return &TimeView{Time: o}
	})
}
