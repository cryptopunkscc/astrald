package file

import (
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/components/storage"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type fileMapperStorage struct {
	dir string
}

type fileMapper struct {
	dir  string
	name string
}

type fileMapperReader struct {
	*os.File
}

func (f *fileMapperStorage) Mapper() (storage.FileMapper, error) {
	return &fileMapper{dir: f.dir}, nil
}

func (f *fileMapperStorage) Reader(name string) (storage.FileReader, error) {
	path := filepath.Join(f.dir, name)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	buff, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	path = string(buff)
	file, err = os.Open(path)
	return &fileMapperReader{File: file}, nil
}

func (f *fileMapperStorage) List() (names []string, err error) {
	return listNames(f.dir)
}

func (f *fileMapper) Map(path string) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	_ = src.Close()
	dst, err := ioutil.TempFile(f.dir, "tmp-")
	if err != nil {
		return err
	}
	defer dst.Close()
	f.name = dst.Name()
	_, err = sio.NewWriter(dst).WriteString(path)
	if err != nil {
		return err
	}
	err = dst.Sync()
	if err != nil {
		return err
	}
	return err
}

func (f *fileMapper) Rename(name string) error {
	dstPath := filepath.Join(f.dir, name)
	return rename(f.name, dstPath)
}

func (f *fileMapperReader) Size() (int64, error) {
	return fileSize(f.File)
}
