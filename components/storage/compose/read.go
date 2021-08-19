package compose

import "github.com/cryptopunkscc/astrald/components/storage"

type composeReadStorage struct {
	arr []storage.ReadStorage
}

func (c composeReadStorage) Reader(
	name string,
) (reader storage.FileReader, err error) {
	for _, s := range c.arr {
		reader, err = s.Reader(name)
		if err == nil {
			break
		}
	}
	return
}

func (c composeReadStorage) List() ([]string, error) {
	var acc []string
	for _, s := range c.arr {
		list, err := s.List()
		if err != nil {
			return nil, err
		}
		acc = append(acc, list...)
	}
	return acc, nil
}
