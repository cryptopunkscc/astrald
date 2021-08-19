package apphost

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
)

// Generic, system-wide directory used as a fallback when no other setting exists
const defaultAstralDir = "/var/run/astrald"

func serveUnix(ctx context.Context) (<-chan net.Conn, error) {
	outCh := make(chan net.Conn)

	listener, err := makeControlSocket()
	if err != nil {
		return nil, err
	}

	go func(listener net.Listener) {
		defer close(outCh)

		go func() {
			<-ctx.Done()
			listener.Close()
		}()

		addr := listener.Addr().String()

		log.Println("apps: listen unix", addr)

		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}
			outCh <- conn
		}

		log.Println("apps: closed unix", addr)
	}(listener)

	return outCh, nil
}

// makeSocket creates a new unix socket and returns its listener. If no name is provided a random one is used.
func makeControlSocket() (net.Listener, error) {
	fullPath := filepath.Join(astralDir(), unixSocketName)

	if fi, err := os.Stat(fullPath); err == nil {
		if !fi.Mode().IsRegular() {
			log.Println("apps: stale unix socket found, removing...")
			err := os.Remove(fullPath)
			if err != nil {
				return nil, err
			}
		}
	}

	listen, err := net.Listen("unix", fullPath)
	if err != nil {
		return nil, err
	}

	return listen, nil
}

// TODO: This should be injected by the node
func astralDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return defaultAstralDir
	}

	dir := filepath.Join(cfgDir, "astrald")
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		fmt.Println("astrald dir erreror:", err)
		return defaultAstralDir
	}

	return dir
}
