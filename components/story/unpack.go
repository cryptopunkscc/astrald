package story

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/serializer"
	"io"
	"log"
)

func UnpackBytes(story []byte) (*Story, error) {
	return Unpack(bytes.NewReader(story))
}

func Unpack(reader io.Reader) (story *Story, err error) {
	p := serializer.NewReader(reader)

	// validate magic bytes
	magic, err := p.ReadN(LenMagicBytes)
	if err != nil {
		return
	}
	if !bytes.Equal(magic, MagicBytes[:]) {
		err = errors.New("cannot unpack story: invalid magic number: " + string(magic) + ", len: " + string(len(magic)))
		return
	}

	// init story
	story = &Story{}

	// read timestamp
	story.timestamp, err = p.ReadUint64()
	if err != nil {
		return
	}

	// read sizes
	typeSize, err := p.ReadByte()
	if err != nil {
		return
	}
	authorSize, err := p.ReadByte()
	if err != nil {
		return
	}
	refsCount, err := p.ReadByte()
	if err != nil {
		return
	}
	dataSize, err := p.ReadUint16()
	if err != nil {
		return
	}

	// values
	story.typ, err = p.ReadN(int(typeSize))
	if err != nil {
		return
	}
	story.author, err = p.ReadN(int(authorSize))
	if err != nil {
		return
	}
	story.refs, err = readRefs(p, refsCount)
	if err != nil {
		return
	}

	if dataSize == 0 {
		pac := story.Pack()
		log.Println("unpacked story size:", len(pac))
		return
	}
	story.data, err = p.ReadN(int(dataSize))
	if err != nil {
		return
	}

	return
}

func readRefs(p io.Reader, size byte) (refs []fid.ID, err error) {
	var refsBuff [fid.Size]byte
	for i := 0; i < int(size); i++ {
		_, err = p.Read(refsBuff[:])
		if err != nil {
			return
		}
		id := fid.Unpack(refsBuff)
		refs = append(refs, id)
	}
	return
}
