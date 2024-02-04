package content

import "github.com/cryptopunkscc/astrald/auth/id"

type Descriptor struct {
	Source id.Identity
	Data   DescriptorData
}

type DescriptorData interface {
	DescriptorType() string
}

type TypeDescriptor struct {
	Method string
	Type   string
}

func (TypeDescriptor) DescriptorType() string {
	return "mod.content.type"
}

type LabelDescriptor struct {
	Label string
}

func (LabelDescriptor) DescriptorType() string {
	return "mod.content.label"
}
