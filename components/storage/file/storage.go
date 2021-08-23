package file

import (
	"github.com/cryptopunkscc/astrald/components/storage"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type fileStorage struct {
	dir string
}

type fileReader struct {
	*os.File
}

type fileWriter struct {
	dir string
	*os.File
}

func (fs fileStorage) Reader(name string) (storage.FileReader, error) {
	path := filepath.Join(fs.dir, name)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return fileReader{File: file}, nil
}

func (fs fileStorage) Writer() (storage.FileWriter, error) {
	file, err := ioutil.TempFile(fs.dir, "tmp-")
	if err != nil {
		log.Println("Cannot create tmp file", err)
		return nil, err
	}
	return fileWriter{
		File: file,
		dir:  fs.dir,
	}, nil
}

func (fs fileStorage) List() (names []string, err error) {
	return ListNames(fs.dir)
}

func (f fileReader) Size() (int64, error) {
	return fileSize(f.File)
}

func (f fileWriter) Rename(name string) error {
	dstPath := filepath.Join(f.dir, name)
	return rename(f.File.Name(), dstPath)
}
