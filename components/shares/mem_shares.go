package shares

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
)

type memShares struct {
	shares map[api.Identity]map[fid.ID]struct{}
}

func NewMemShares() Shares {
	return &memShares{
		shares: map[api.Identity]map[fid.ID]struct{}{},
	}
}

func (m *memShares) Add(id api.Identity, file fid.ID) error {
	files, contains := m.shares[id]
	if !contains {
		files = map[fid.ID]struct{}{}
		m.shares[id] = files
	}
	files[file] = struct{}{}
	return nil
}

func (m *memShares) Remove(id api.Identity, file fid.ID) error {
	files, contains := m.shares[id]
	if contains {
		delete(files, file)
		if len(files) == 0 {
			delete(m.shares, id)
		}
	}
	return nil
}

func (m *memShares) List(id api.Identity) ([]fid.ID, error) {
	filesSet, contains := m.shares[id]
	if !contains {
		return []fid.ID{}, nil
	}
	files := make([]fid.ID, len(filesSet))
	index := 0
	for file := range filesSet {
		files[index] = file
		index++
	}
	return files, nil
}

func (m *memShares) Contains(id api.Identity, file fid.ID) (bool, error) {
	files, contains := m.shares[id]
	if !contains {
		return false, nil
	}
	_, contains = files[file]
	return contains, nil
}
