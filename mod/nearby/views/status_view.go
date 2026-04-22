package views

import (
	"strconv"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
	"github.com/cryptopunkscc/astrald/mod/nearby"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type StatusView struct {
	*nearby.Status
}

func (v StatusView) Render() string {
	s := log.DefaultViewer.Render(v.Identity)
	l := len(v.Attachments.Objects())
	if l > 0 {
		var first = true
		s += " ("
		for _, a := range v.attachments() {
			if !first {
				s += ", "
			}
			s += a
			first = false
		}
		s += ")"
	}

	return s
}

func init() {
	log.DefaultViewer.Set(nearby.Status{}.ObjectType(), func(o astral.Object) astral.Object {
		return StatusView{o.(*nearby.Status)}
	})
}

func (v StatusView) attachments() (list []string) {
	var more int
	var alias string
	var flags []string
	var endpoints int

	for _, attachment := range v.Attachments.Objects() {
		switch a := attachment.(type) {
		case *dir.Alias:
			alias = a.String()
		case *nearby.Flag:
			flags = append(flags, a.String())
		case *nodes.EndpointWithTTL:
			endpoints++
		default:
			more++
		}
	}

	if len(alias) > 0 {
		list = append(list, styles.GreenText.Render(alias))
	}

	if len(flags) > 0 {
		for _, flag := range flags {
			list = append(list, styles.YellowText.Render(flag))
		}
	}

	if endpoints > 0 {
		item := styles.WhiteText.Render(strconv.Itoa(endpoints))
		item += styles.GrayText.Render(" endpoints")
		list = append(list, item)
	}

	if more > 0 {
		item := styles.WhiteText.Render(strconv.Itoa(more))
		item += styles.GrayText.Render(" other attachment")
		list = append(list, item)
	}

	return
}
