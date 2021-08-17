package file

import (
	"errors"
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

func (f fileReader) Size() (int64, error) {
	stat, err := f.Stat()
	if err != nil {
		return -1, err
	}
	return stat.Size(), nil
}

type fileWriter struct {
	*os.File
	dir string
}

func NewStorage(astralHome string) storage.Storage {
	storageDir := filepath.Join(astralHome, "storage")
	err := os.Mkdir(storageDir, 0700)
	log.Println(err)
	return fileStorage{dir: storageDir}
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

func (f fileWriter) Rename(name string) error {
	dstPath := filepath.Join(f.dir, name)
	_, err := os.Stat(dstPath)
	if !errors.Is(err, os.ErrNotExist) {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return errors.New("file already exist")
	}
	err = os.Rename(f.Name(), dstPath)
	if err != nil {
		log.Println("cannot rename file", err)
		return err
	}
	return nil
}
