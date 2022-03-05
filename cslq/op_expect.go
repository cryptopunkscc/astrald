package cslq

import (
	"bytes"
	"errors"
	"io"
)

type OpExpect []Op

func (op OpExpect) Encode(w io.Writer, data *Fifo) error {
	for _, o := range op {
		if err := o.Encode(w, data); err != nil {
			return err
		}
	}
	return nil
}

func (op OpExpect) Decode(r io.Reader, data *Fifo) error {
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
