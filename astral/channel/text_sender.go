package channel

import (
	"encoding"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type TextSender struct {
	Base64 bool // force base64
	w      io.Writer
}

var _ Sender = &TextSender{}

func NewTextSender(w io.Writer) *TextSender {
	return &TextSender{w: w}
}

func (sender TextSender) Send(obj astral.Object) (err error) {
	// write the type
	_, err = fmt.Fprintf(sender.w, "#[%s]", obj.ObjectType())
	if err != nil {
		return
	}

	// check if the object is a TextMarshaler
	m, ok := obj.(encoding.TextMarshaler)
	if ok && !sender.Base64 {
		var text []byte

		// marshal the object into text
		text, err = m.MarshalText()
		if err != nil {
			return err
		}

		_, err = sender.w.Write([]byte(" " + string(text) + "\n"))
		return
	}

	_, err = sender.w.Write([]byte(":"))
	if err != nil {
		return
	}

	enc := base64.NewEncoder(base64.StdEncoding, sender.w)
	_, err = obj.WriteTo(enc)
	enc.Close()
	if err != nil {
		return
	}

	_, err = sender.w.Write([]byte("\n"))
	return err
}
