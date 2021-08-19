package compose

import "github.com/cryptopunkscc/astrald/components/storage"

func NewReadStorage(
	s ...storage.ReadStorage,
) storage.ReadStorage {
	return composeReadStorage{s}
}
