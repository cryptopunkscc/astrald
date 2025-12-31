package channel

import (
	"bytes"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type BinaryReader struct {
	bp *astral.Blueprints
	r  io.Reader
}

var _ Reader = &BinaryReader{}

func NewBinaryReader(r io.Reader) *BinaryReader {
	return &BinaryReader{r: r, bp: astral.ExtractBlueprints(r)}
}

func (b BinaryReader) Read() (object astral.Object, err error) {
	var frame astral.Bytes16

	_, err = frame.ReadFrom(b.r)
	if err != nil {
		return
	}

	object, _, err = b.bp.Read(bytes.NewReader(frame))

	return
}
