package serialize

import "io"

type Serializer struct {
	*Parser
	*Formatter
}

func NewSerializer(rw io.ReadWriter) Serializer {

	return Serializer{
		Parser:    NewParser(rw),
		Formatter: NewFormatter(rw),
	}
}
