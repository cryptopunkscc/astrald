package file

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/shares"
	"github.com/cryptopunkscc/astrald/components/storage/file"
	"os"
	"path"
)

type sharedFiles struct {
	dir string
}

func NewSharedFiles(home string) shares.Shared {
	dir, err := file.ResolveDir(home, "shared")
	if err != nil {
		panic(err)
	}
	return &sharedFiles{dir}
}

func (m *sharedFiles) Add(id api.Identity, file fid.ID) error {
	var added []fid.ID
	filePath := path.Join(m.dir, string(id))

	// Try to open existing file
	f, err := os.Open(filePath)
	if err == nil {
		// Read added files
		for err == nil {
			var currId fid.ID
			if currId, _, err = fid.Read(f); err != nil {
				added = append(added, currId)
			} else {
				_ = f.Close()
			}
		}

		// Check if is already added
		for _, currId := range added {
			if currId.String() == file.String() {
				return nil
			}
		}
	}

	// Append to file
	f, err = os.Create(filePath)
	added = append(added, file)
	for _, fileId := range added {
		err := fileId.Write(f)
		if err != nil {
			return err
		}
	}
	return f.Sync()
}

func (m *sharedFiles) Remove(id api.Identity, file fid.ID) error {
	var added []fid.ID
	filePath := path.Join(m.dir, string(id))

	// Try to open existing file
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	// Filter added files
	for err == nil {
		var currId fid.ID
		if currId, _, err = fid.Read(f); err != nil {
			if currId.String() != file.String() {
				added = append(added, currId)
			}
		} else {
			_ = f.Close()
		}
	}

	// Append to file
	f, err = os.Create(filePath)
	added = append(added, file)
	for _, fileId := range added {
		err := fileId.Write(f)
		if err != nil {
			return err
		}
	}
	return f.Sync()
}

func (m *sharedFiles) List(id api.Identity) ([]fid.ID, error) {
	var added []fid.ID
	filePath := path.Join(m.dir, string(id))

	// Try to open existing file
	f, err := os.Open(filePath)
	if err != nil {
		return added, nil
	}

	// Filter added files
	for err == nil {
		var currId fid.ID
		if currId, _, err = fid.Read(f); err != nil {
			added = append(added, currId)
		} else {
			_ = f.Close()
		}
	}

	return added, nil
}

func (m *sharedFiles) Contains(id api.Identity, file fid.ID) (bool, error) {
	filePath := path.Join(m.dir, string(id))

	// Try to open existing file
	f, err := os.Open(filePath)
	if err != nil {
		return false, nil
	}
	defer func() { _ = f.Close() }()

	// Search for file id
	for err == nil {
		currId, _, err := fid.Read(f)
		if err != nil {
			return false, err
		}
		if currId.String() == file.String() {
			return true, nil
		}
	}
	return false, nil
}
