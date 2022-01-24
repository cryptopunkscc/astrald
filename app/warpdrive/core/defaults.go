package core

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"os"
	"path/filepath"
)

func newDefaultResolver() api.Resolver {
	return newFsResolver(userFilesStorage())
}

func receivedFilesStorage() storage {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(home, "warpdrive", "received")
	err = os.MkdirAll(dir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	return newStorage(dir)
}

func userFilesStorage() storage {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return newStorage(dir)
}

func filesStorage() storage {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("error fetching user's config dir:", err)
		os.Exit(0)
	}

	dir := filepath.Join(cfgDir, "warpdrive")
	os.MkdirAll(dir, 0700)

	return newStorage(dir)
}
