package views

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type TagView struct {
	*log.Tag
}

func (v TagView) Render() string {
	c := styles.TextColorFromString(v.Tag.String())
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(c))

	return "[" + style.Render(v.Tag.String()) + "] "

}

func init() {
	log.DefaultViewer.Set(log.Tag(0).ObjectType(), func(o astral.Object) astral.Object {
		return &TagView{o.(*log.Tag)}
	})
}
