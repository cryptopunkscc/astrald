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

func (fs *fileMapperStorage) Mapper() (storage.FileMapper, error) {
	return &fileMapper{dir: fs.dir}, nil
}

func (fs *fileMapperStorage) Reader(name string) (storage.FileReader, error) {
	path := filepath.Join(fs.dir, name)
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

func (fs *fileMapperStorage) List() (names []string, err error) {
	return listNames(fs.dir)
}

func (fs *fileMapper) Map(path string) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	_ = src.Close()
	dst, err := ioutil.TempFile(fs.dir, "tmp-")
	_, err = sio.NewWriter(dst).WriteString(path)
	if err != nil {
		return err
	}
	return err
}

func (fs *fileMapper) Rename(name string) error {
	return rename(fs.name, name)
}

func (f *fileMapperReader) Size() (int64, error) {
	return fileSize(f.File)
}
