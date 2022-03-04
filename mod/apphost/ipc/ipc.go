package ipc

import (
	"errors"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Dial(target string) (net.Conn, error) {
	parts := strings.SplitN(target, ":", 2)
	proto, addr := parts[0], parts[1]

	switch proto {
	case "tcp":
		return net.Dial("tcp", addr)
	case "unix":
		return net.Dial("unix", addr)
	default:
		return nil, errors.New("unsupported protocol")
	}
}

func Listen(protocol string) (net.Listener, error) {
	switch protocol {
	case "tcp":
		return net.Listen("tcp", "127.0.0.1:0")
	case "unix":
		return net.Listen(
			"unix",
			filepath.Join(
				os.TempDir(),
				"apphostclient."+tempName(10),
			),
		)
	default:
		return nil, errors.New("unsupported protocol")
	}
}

func tempName(length int) (s string) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789_"
	var name = make([]byte, length)
	for i := 0; i < len(name); i++ {
		name[i] = charset[rand.Intn(len(charset))]
	}
	return string(name[:])
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
