package log

var _ Output = &OutputSplitter{}

type OutputSplitter struct {
	Outputs []Output
}

func NewOutputSplitter(printers ...Output) *OutputSplitter {
	return &OutputSplitter{Outputs: printers}
}

func (out *OutputSplitter) Do(ops ...Op) {
	for _, o := range out.Outputs {
		o.Do(ops...)
	}
}
