package objects

import (
	"bytes"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

func (mod *Module) Decode(data []byte) (objects.Object, error) {
	var r = bytes.NewReader(data)

	_, obj, err := mod.decodeStream(r)

	return obj, err
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
	err := cslq.Encode(buf, "vv", astral.ObjectHeader(obj.ObjectType()), obj)
	return buf.Bytes(), err
}

func (mod *Module) decodeStream(r io.Reader) (object.ID, objects.Object, error) {
	var err error
	var header astral.ObjectHeader

	rr := object.NewReadResolver(r)

	err = cslq.Decode(rr, "v", &header)
	if err != nil {
		return object.ID{}, nil, err
	}

	decoder, found := mod.decoders.Get(header.String())
	if !found {
		return object.ID{}, nil, fmt.Errorf("decoder for %s not found", header.String())
	}

	objectBytes, err := io.ReadAll(rr)
	if err != nil {
		return object.ID{}, nil, err
	}

	objectID := rr.Resolve()

	object, err := decoder(objectBytes)

	return objectID, object, err
}
