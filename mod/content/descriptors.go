package content

type TypeDesc struct {
	Method      string
	ContentType string
}

func (TypeDesc) Type() string {
	return "mod.content.type"
}
func (d TypeDesc) String() string { return d.ContentType }

type LabelDesc struct {
	Label string
}

func (LabelDesc) Type() string {
	return "mod.content.label"
}
func (d LabelDesc) String() string { return d.Label }
