package scanner

import (
	"os"
	"path/filepath"
	"time"
)

type HandlePath func(path string, modTime time.Time)

func Scan(path string, handle HandlePath) error {
	return filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			handle(path, info.ModTime())
			return nil
		})
}
