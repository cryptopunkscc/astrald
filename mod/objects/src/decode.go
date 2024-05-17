package objects

import (
	"bytes"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"io"
)

func (mod *Module) Decode(data []byte) (objects.Object, error) {
	var r = bytes.NewReader(data)

	return mod.decodeStream(r)
}

func (mod *Module) SetDecoder(s string, decoder objects.Decoder) error {
	if decoder == nil {
		mod.decoders.Delete(s)
	} else {
		mod.decoders.Replace(s, decoder)
	}
	return nil
}

func (mod *Module) Encode(obj objects.Object) ([]byte, error) {
	var buf = &bytes.Buffer{}
	err := cslq.Encode(buf, "vv", adc.Header(obj.ObjectType()), obj)
	return buf.Bytes(), err
}

func (mod *Module) decodeStream(r io.Reader) (objects.Object, error) {
	var err error
	var header adc.Header

	err = cslq.Decode(r, "v", &header)
	if err != nil {
		return nil, err
	}

	decoder, found := mod.decoders.Get(header.String())
	if !found {
		return nil, fmt.Errorf("decoder for %s not found", header.String())
	}

	obj, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return decoder(obj)
}
