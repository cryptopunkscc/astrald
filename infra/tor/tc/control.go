package tc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
)

type Control struct {
	Config
	protocolInfo *ProtocolInfo
	proto        *textproto.Conn
	closeCh      chan struct{}
}

type Onion struct {
	ServiceID  string
	PrivateKey string
}

type watch struct {
	io.ReadWriteCloser
	closeCh chan struct{}
}

func (w watch) Read(p []byte) (int, error) {
	n, err := w.ReadWriteCloser.Read(p)
	if err != nil {
		if w.closeCh != nil {
			close(w.closeCh)
			w.closeCh = nil
		}
	}
	return n, err
}

func watchClosed(rwc io.ReadWriteCloser) (io.ReadWriteCloser, chan struct{}) {
	w := &watch{
		ReadWriteCloser: rwc,
		closeCh:         make(chan struct{}),
	}

	return w, w.closeCh
}

// Connect establishes a new connection to the Tor daemon
func Connect(cfg Config) (*Control, error) {
	conn, err := net.Dial("tcp", cfg.getContolAddr())
	if err != nil {
		return nil, err
	}

	c, ch := watchClosed(conn)

	return &Control{
		Config:  cfg,
		proto:   textproto.NewConn(c),
		closeCh: ch,
	}, nil
}

// Open connects to the daemon and authenticates
func Open(cfg Config) (*Control, error) {
	ctl, err := Connect(cfg)
	if err != nil {
		return nil, err
	}

	if err := ctl.Authenticate(); err != nil {
		ctl.proto.Close()
		return nil, err
	}

	return ctl, nil
}

func (ctl *Control) GetConf(key string) (string, error) {
	_, lines, err := ctl.request("GETCONF %s", key)
	if err != nil {
		return "", err
	}

	for _, l := range lines {
		kv := strings.SplitN(l, "=", 2)
		if len(kv) > 1 {
			return kv[1], nil
		}
	}

	return "", errors.New("no value set")
}

func (ctl *Control) GetInfo(arg string) (string, error) {
	_, lines, err := ctl.request("GETINFO %s", arg)
	if err != nil {
		return "", err
	}

	for _, l := range lines {
		fmt.Println(l)
	}

	for _, l := range lines {
		kv := strings.SplitN(l, "=", 2)
		if len(kv) > 1 {
			return kv[1], nil
		}
	}

	return "", errors.New("no value set")
}

func (ctl *Control) AddOnion(privateKey string, ports map[int]string) (Onion, error) {
	portPart := portsToString(ports)

	_, lines, err := ctl.request("ADD_ONION %s %s", privateKey, portPart)
	if err != nil {
		return Onion{}, err
	}

	vars := parseVarMap(lines)

	return Onion{
		ServiceID:  vars.GetString("ServiceID"),
		PrivateKey: vars.GetString("PrivateKey"),
	}, nil
}

func (ctl *Control) DelOnion(serviceID string) error {
	_, _, err := ctl.request("DEL_ONION %s", serviceID)
	return err
}

func (ctl *Control) WaitClose() <-chan struct{} {
	return ctl.closeCh
}

func (ctl *Control) Close() error {
	return ctl.proto.Close()
}

func (ctl *Control) request(format string, args ...interface{}) (int, []string, error) {
	id, err := ctl.proto.Cmd(format, args...)
	if err != nil {
		return 0, []string{}, nil
	}

	ctl.proto.StartResponse(id)
	defer ctl.proto.EndResponse(id)

	code, body, err := ctl.proto.ReadResponse(statusCodeOK)
	if err != nil {
		return 0, nil, err
	}

	lines := strings.Split(body, "\n")
	if len(lines) > 1 {
		if !strings.Contains(lines[0], "=") {
			lines = lines[1:]
		}
		if lines[len(lines)-1] == "OK" {
			lines = lines[:len(lines)-1]
		}
	}

	return code, lines, err
}

func Port(port int, addr string) map[int]string {
	return map[int]string{port: addr}
}

func portsToString(ports map[int]string) string {
	if ports == nil {
		return ""
	}

	list := make([]string, 0)
	for k, v := range ports {
		list = append(list, fmt.Sprintf("Port=%d,%s", k, v))
	}
	return strings.Join(list, " ")
}
