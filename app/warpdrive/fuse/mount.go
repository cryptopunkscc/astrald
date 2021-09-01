package fuse

import (
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
	"log"
	"os"
	"path/filepath"
)

func MountWarpDrive() {
	mountDir := "warpdrive"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}

	mountPoint := filepath.Join(homeDir, mountDir)

	_ = os.Mkdir(mountPoint, 0777)

	options := &fs.Options{
		UID: uint32(os.Getuid()),
		GID: uint32(os.Getgid()),
	}

	server, err := fs.Mount(mountPoint, &peersDir{}, options)
	if err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}

	log.Println("Mounted!")
	server.Wait()
}
