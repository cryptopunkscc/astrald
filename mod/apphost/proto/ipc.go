package proto

import (
	"github.com/akutz/memconn"
	"github.com/cryptopunkscc/astrald/auth/id"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EnvKeyAddr is the name of the environment variable which contains a semicolon-separated list of apphost addresses.
const EnvKeyAddr = "ASTRALD_APPHOST_ADDR"

// EnvKeyToken is the name of the environment variable which contains the apphost access token.
const EnvKeyToken = "ASTRALD_APPHOST_TOKEN"

func Dial(target string) (c net.Conn, err error) {
	parts := strings.SplitN(target, ":", 2)
	proto, addr := parts[0], parts[1]

	switch proto {
	case "tcp":
		c, err = net.Dial("tcp", addr)

	case "unix":
		c, err = net.Dial("unix", addr)

	case "memu", "memb":
		c, err = memconn.Dial(proto, addr)

	default:
		err = ErrUnsupportedProtocol
	}

	return
}

func Listen(ipcAddress string) (net.Listener, error) {
	var protocol, address string

	if parts := strings.SplitN(ipcAddress, ":", 2); len(parts) < 2 {
		return nil, ErrInvalidIPCAddress
	} else {
		protocol, address = parts[0], parts[1]
	}

	switch protocol {
	case "tcp":
		return net.Listen("tcp", address)

	case "unix":
		var path = address

		if strings.HasPrefix(path, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				path = filepath.Join(home, path[2:])
			}
		}

		return net.Listen("unix", path)

	case "memu", "memb":
		return memconn.Listen(protocol, address)

	default:
		return nil, ErrUnsupportedProtocol
	}
}

func ListenAny(protocol string) (net.Listener, error) {
	switch protocol {
	case "tcp":
		return net.Listen("tcp", "127.0.0.1:0")

	case "unix":
		return net.Listen(
			"unix",
			filepath.Join(
				os.TempDir(),
				"apphostclient."+tempName(16),
			),
		)

	case "memu", "memb":
		identity, err := id.GenerateIdentity()
		if err != nil {
			return nil, err
		}
		return memconn.Listen("memu", identity.String())

	default:
		return nil, ErrUnsupportedProtocol
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
