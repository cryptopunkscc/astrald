package tor

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/textproto"
	"strconv"
	"strings"
)

const controlAddr = "127.0.0.1:9051"
const OK = 250

type Control struct {
	*textproto.Conn
	auth map[string]string
	ver  map[string]string
}

func Open() (*Control, error) {
	conn, err := textproto.Dial("tcp", controlAddr)
	if err != nil {
		log.Println("tor error:", err)
		return nil, err
	}

	ctl := &Control{Conn: conn}

	err = ctl.readProtocolInfo()
	if err != nil {
		log.Println("tor error:", err)
		return nil, err
	}

	err = ctl.authenticate()
	if err != nil {
		log.Println("tor error:", err)
		return nil, err
	}

	log.Println("tor version", ctl.version())

	return ctl, nil
}

func (ctl *Control) GetInfo(arg string) (string, error) {
	_, lines, err := ctl.request("GETINFO %s", arg)
	if err != nil {
		return "", err
	}

	return strings.SplitN(lines[0], "=", 2)[1], nil
}

func (ctl *Control) NewService(ports map[int]string) (*Service, error) {
	return ctl.StartService("", ports)
}

func (ctl *Control) StartService(privateKey string, ports map[int]string) (*Service, error) {
	if privateKey == "" {
		privateKey = "NEW:BEST"
	}

	list := make([]string, 0)
	for k, v := range ports {
		list = append(list, fmt.Sprintf("Port=%d,%s", k, v))
	}
	portPart := strings.Join(list, " ")

	_, lines, err := ctl.request("ADD_ONION %s %s", privateKey, portPart)
	if err != nil {
		log.Println("tor error:", err)
		return nil, err
	}

	vars := varsToMap(lines)

	return &Service{
		ctl:        ctl,
		serviceID:  vars["ServiceID"],
		privateKey: vars["PrivateKey"],
	}, nil
}

func (ctl *Control) stopService(serviceID string) error {
	_, _, err := ctl.request("DEL_ONION %s", serviceID)
	return err
}

func (ctl *Control) request(format string, args ...interface{}) (int, []string, error) {
	id, err := ctl.Cmd(format, args...)
	if err != nil {
		return 0, []string{}, nil
	}

	ctl.StartResponse(id)
	defer ctl.EndResponse(id)

	code, body, err := ctl.ReadResponse(OK)
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

func (ctl *Control) readProtocolInfo() error {
	_, lines, err := ctl.request("PROTOCOLINFO 1")
	if err != nil {
		return err
	}

	for _, line := range lines {
		words := strings.Split(line, " ")
		scope, values := words[0], varsToMap(words[1:])

		switch scope {
		case "AUTH":
			ctl.auth = values
		case "VERSION":
			ctl.ver = values
		}
	}
	return nil
}

func (ctl *Control) authenticate() error {
	cookie, err := ctl.authCookie()
	if err != nil {
		return err
	}

	_, _, err = ctl.request("AUTHENTICATE %s", cookie)
	if err != nil {
		return err
	}

	return err
}

func (ctl *Control) authCookie() (string, error) {
	bytes, err := ioutil.ReadFile(ctl.cookieFile())
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func (ctl *Control) cookieFile() string {
	if ctl.auth == nil {
		return ""
	}

	if s, err := strconv.Unquote(ctl.auth["COOKIEFILE"]); err == nil {
		return s
	}

	return ctl.auth["COOKIEFILE"]
}

func (ctl *Control) authMethods() []string {
	if ctl.auth == nil {
		return []string{}
	}

	return strings.Split(ctl.auth["METHODS"], ",")
}

func (ctl *Control) version() string {
	if ctl.ver == nil {
		return ""
	}

	if s, err := strconv.Unquote(ctl.ver["Tor"]); err == nil {
		return s
	}

	return ctl.ver["Tor"]
}

func varsToMap(words []string) map[string]string {
	m := make(map[string]string)
	for _, w := range words {
		kv := strings.SplitN(w, "=", 2)
		m[kv[0]] = kv[1]
	}
	return m
}
