package main

import (
	"errors"
	"fmt"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	_store "github.com/cryptopunkscc/astrald/proto/store"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Server struct {
	dataDir string
}

func (server *Server) Run() error {
	log.Println("storaged starting...")

	port, err := astral.Register("storage")
	if err != nil {
		return err
	}

	log.Println("data dir:", server.dataDir)

	store := &DirStore{dataDir: server.dataDir}

	for conn := range port.AcceptAll() {
		conn := conn
		go func() {
			if err := _store.Serve(conn, store); err != nil {
				if !errors.Is(err, io.EOF) {
					log.Println("error:", err)
				}
			}
			conn.Close()
		}()
	}

	return nil
}

func main() {
	var dataDir = "./storage"

	if home, err := os.UserHomeDir(); err == nil {
		dataDir = filepath.Join(home, ".config/astrald/storage")
	}

	if len(os.Args) >= 2 {
		dataDir = os.Args[1]
	}

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		fmt.Println("init error:", err)
		os.Exit(1)
	}

	server := &Server{dataDir: dataDir}

	if err := server.Run(); err != nil {
		fmt.Println("server error:", err)
		os.Exit(2)
	}
}
