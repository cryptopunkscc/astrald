package term

import "io"

var _ Renderer = &BasicTerminal{}

type BasicTerminal struct {
	w io.Writer
}

func NewBasicTerminal(w io.Writer) *BasicTerminal {
	return &BasicTerminal{
		w: w,
	}
}

func (t *BasicTerminal) Text(text string) {
	t.w.Write([]byte(text))
}

func (t *BasicTerminal) SetColor(color Color) {
}
