package cslq

import (
	"bytes"
	"errors"
	"io"
)

type OpConst []Op

func (op OpConst) Encode(w io.Writer, data *Fifo) error {
	for _, o := range op {
		if err := o.Encode(w, data); err != nil {
			return err
		}
	}
	return nil
}

func (op OpConst) Decode(r io.Reader, data *Fifo) error {
	for _, o := range op {
		var buf = &bytes.Buffer{}
		if err := o.Encode(buf, data); err != nil {
			return err
		}

		expected := buf.Bytes()

		if len(expected) == 0 {
			continue
		}

		var actual = make([]byte, len(expected))
		if _, err := io.ReadFull(r, actual); err != nil {
			return err
		}

		if bytes.Compare(expected, actual) != 0 {
			return errors.New("unexpected data")
		}
	}

	return nil
}

func (op OpConst) String() string {
	var s = "<"
	for _, sub := range op {
		s = s + sub.String()
	}
	return s + ">"
}
