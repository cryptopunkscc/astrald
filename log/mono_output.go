package log

import "io"

var _ Output = &MonoOutput{}

type MonoOutput struct {
	io.Writer
}

func NewMonoOutput(writer io.Writer) *MonoOutput {
	return &MonoOutput{Writer: writer}
}

func (out *MonoOutput) Do(ops ...Op) {
	for _, op := range ops {
		switch op := op.(type) {
		case OpText:
			out.Write([]byte(op.Text))
		}
	}
}
