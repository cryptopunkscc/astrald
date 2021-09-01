package file

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func ResolveDir(root string, name string) (string, error) {
	storageDir := filepath.Join(root, name)
	stat, err := os.Stat(storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(storageDir, 0700)
			if err != nil {
				return "", errors.New("cannot create storage: " + err.Error())
			}
			log.Println("created storage dir in", storageDir)
		} else if stat.Mode().Perm() != 0700 {
			err := os.Chmod(storageDir, 0700)
			if err != nil {
				return "", errors.New("cannot change storage mode: " + err.Error())
			}
		}
	}
	return storageDir, nil
}

func ListNames(dir string) (names []string, err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	names = make([]string, len(files))
	for i, file := range files {
		names[i] = file.Name()
	}
	return
}

func fileSize(f *os.File) (int64, error) {
	stat, err := f.Stat()
	if err != nil {
		return -1, err
	}
	return stat.Size(), nil
}

func rename(oldPath, newPath string) error {
	_, err := os.Stat(newPath)
	if !errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(oldPath)
		return errors.New("file already exist")
	}
	err = os.Rename(oldPath, newPath)
	if err != nil {
		log.Println("cannot rename file", oldPath, newPath, err)
		return err
	}
	return nil
}
