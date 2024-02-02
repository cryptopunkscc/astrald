package content

import "github.com/cryptopunkscc/astrald/auth/id"

type Descriptor struct {
	Source id.Identity
	Info   Info
}

type Info interface {
	InfoType() string
}

type TypeDescriptor struct {
	Method string
	Type   string
}

func (TypeDescriptor) InfoType() string {
	return "mod.content.type"
}

type LabelDescriptor struct {
	Label string
}

func (LabelDescriptor) InfoType() string {
	return "mod.content.label"
}
