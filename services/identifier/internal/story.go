package internal

import (
	_bytes "bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/components/story"
)

const storyMimeType = "application/lore"

func GetStoryType(bytes []byte) (string, error) {
	magicBytes := bytes[0:4]
	if _bytes.Equal(magicBytes, story.MagicBytes[:]) {
		return storyMimeType, nil
	}
	return "", errors.New("not story type")
}
