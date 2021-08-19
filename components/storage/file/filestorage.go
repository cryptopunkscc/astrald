package file

import (
	"errors"
	"github.com/cryptopunkscc/astrald/components/serializer"
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
	serializer.Parser
	*os.File
}

type fileWriter struct {
	*os.File
	serializer.Formatter
	dir string
}

func NewStorage(astralHome string) storage.Storage {
	storageDir := filepath.Join(astralHome, "storage")
	stat, err := os.Stat(storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(storageDir, 0700)
			if err != nil {
				panic("cannot create storage: " + err.Error())
			}
			log.Println("created storage dir in", storageDir)
		} else if stat.Mode().Perm() != 0700 {
			err := os.Chmod(storageDir, 0700)
			if err != nil {
				panic("cannot change storage mode: " + err.Error())
			}
		}
	}
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

func (fs fileStorage) List() (names []string, err error) {
	files, err := ioutil.ReadDir(fs.dir)
	if err != nil {
		return
	}
	names = make([]string, len(files))
	for i, file := range files {
		names[i] = file.Name()
	}
	return
}

func (f fileReader) Size() (int64, error) {
	stat, err := f.Stat()
	if err != nil {
		return -1, err
	}
	return stat.Size(), nil
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
