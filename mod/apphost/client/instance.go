package astral

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const defaultAstralDir = "/var/run/astrald"

var defaultInstance = NewAstral(astralDir())

func Instance() *Astral {
	return defaultInstance
}

func Reqister(name string) (*Port, error) {
	return Instance().Register(name)
}

func Query(identity string, query string) (io.ReadWriteCloser, error) {
	return Instance().Query(identity, query)
}

func astralDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return defaultAstralDir
	}

	dir := filepath.Join(cfgDir, "astrald")
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		fmt.Println("astrald dir error:", err)
		return defaultAstralDir
	}

	return dir
}
