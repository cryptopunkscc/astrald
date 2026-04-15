package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

const DefaultTimeLayout = "15:04:05.000"

type TimeView struct {
	*astral.Time
	Layout string
	Style  *lipgloss.Style
}

func NewTimeView(time *astral.Time) *TimeView {
	return &TimeView{Time: time, Layout: DefaultTimeLayout}
}

func NewTimeViewStyled(time *astral.Time, layout string, style lipgloss.Style) *TimeView {
	return &TimeView{Time: time, Layout: layout, Style: &style}
}

func (v TimeView) Render() (out string) {
	layout := v.Layout
	if layout == "" {
		layout = DefaultTimeLayout
	}

	t := v.Time.Time().Format(layout)

	if v.Style != nil {
		return v.Style.Render(t)
	}

	if !strings.HasSuffix(layout, ".000") {
		return styles.GrayText.Render(t)
	}

	l := len(t)

	msec := t[l-4:]
	t = t[0 : l-4]

	return styles.GrayText.Render(t) + styles.DarkGrayText.Render(msec)
}

func init() {
	log.Set(func(o *astral.Time) astral.Object {
		return &TimeView{Time: o}
	})
}
