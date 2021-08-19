package compose

import "github.com/cryptopunkscc/astrald/components/storage"

type composeStorage struct {
	storage.ReadStorage
	storage.WriteStorage
	storage.MapStorage
}

func NewComposeStorage(
	read storage.ReadStorage,
	write storage.WriteStorage,
	mapper storage.MapStorage,
) storage.Storage {
	return composeStorage{
		ReadStorage:  read,
		WriteStorage: write,
		MapStorage:   mapper,
	}
}
