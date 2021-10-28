package astral

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"
)

const ctlSocket = "ctl.sock"
const defaultUseTCP = false
const randomStringCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Astral struct {
	rootDir string
	UseTCP  bool
}

func NewAstral(rootDir string) *Astral {
	return &Astral{rootDir: rootDir, UseTCP: defaultUseTCP}
}

func (astral *Astral) Register(name string) (*Port, error) {
	// connect to the daemon
	daemon, err := astral.dialDaemon()
	if err != nil {
		return nil, err
	}

	ctlConn := proto.NewConn(daemon)

	// set up unix socket
	path := filepath.Join(os.TempDir(), ".astral-"+randomString(32))

	// register
	res, err := ctlConn.Register(name, path)
	if err != nil {
		ctlConn.Close()
		return nil, fmt.Errorf("register error: %v", err)
	}
	if res.Status != proto.StatusOK {
		ctlConn.Close()
		return nil, fmt.Errorf("register error: %v", errors.New(res.Error))
	}

	port, err := NewUnixPort(path, ctlConn)
	if err != nil {
		ctlConn.Close()
		return nil, err
	}

	go func() {
		defer port.Close()

		var buf [1]byte
		_, err := ctlConn.Read(buf[:])
		if err == nil {
			ctlConn.Close()
		}
	}()

	return port, err
}

func (astral *Astral) Query(nodeID string, query string) (io.ReadWriteCloser, error) {
	daemon, err := astral.dialDaemon()
	if err != nil {
		return nil, err
	}

	conn := proto.NewConn(daemon)

	// Send the request
	res, err := conn.Query(nodeID, query)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("query error: %v", err)
	}
	if res.Status != proto.StatusOK {
		conn.Close()
		return nil, fmt.Errorf("query error: %v", errors.New(res.Error))
	}

	return daemon, nil
}

func (astral *Astral) dialDaemon() (net.Conn, error) {
	n, a := "unix", astral.socketPath(ctlSocket)

	if astral.UseTCP {
		n, a = "tcp", "localhost:8625"
	}

	c, err := net.Dial(n, a)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (astral *Astral) socketPath(name string) string {
	return filepath.Join(astral.rootDir, name)
}

func randomString(length int) string {
	s := make([]byte, length)
	for i := range s {
		s[i] = randomStringCharset[rand.Intn(len(randomStringCharset))]
	}
	return string(s)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
