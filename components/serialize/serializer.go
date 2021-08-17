package serialize

import "io"

type Serializer struct {
	io.Closer
	*Parser
	*Formatter
}

func NewSerializer(rw io.ReadWriteCloser) Serializer {

	return Serializer{
		Closer: rw,
		Parser:    NewParser(rw),
		Formatter: NewFormatter(rw),
	}
}
