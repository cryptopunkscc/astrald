package content

type Descriptor interface {
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
