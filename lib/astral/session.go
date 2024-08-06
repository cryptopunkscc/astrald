package astral

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"net"
	"strings"
)

type Session struct {
	*cslq.Endec
	conn  *proto.Conn
	addr  string
	token string
}

func NewSession(conn net.Conn, token string, addr string) *Session {
	return &Session{
		Endec: cslq.NewEndec(conn),
		conn:  proto.NewConn(conn),
		token: token,
		addr:  addr,
	}
}

func (s *Session) Query(remoteID *astral.Identity, query string) (conn *Conn, err error) {
	if err = s.auth(); err != nil {
		s.Close()
		return
	}

	err = s.invoke(proto.CmdQuery, proto.QueryParams{
		Identity: remoteID,
		Query:    query,
	})
	if err != nil {
		s.Close()
		return nil, err
	}

	return &Conn{
		Conn:     s.conn,
		remoteID: remoteID,
		query:    query,
	}, nil
}

func (s *Session) Resolve(name string) (identity *astral.Identity, err error) {
	if err = s.auth(); err != nil {
		return
	}

	err = s.invoke(proto.CmdResolve, proto.ResolveParams{Name: name})
	if err != nil {
		s.Close()
		return
	}

	var data proto.ResolveData

	err = s.Decodef("v", &data)

	return data.Identity, err
}

func (s *Session) NodeInfo(identity *astral.Identity) (info proto.NodeInfoData, err error) {
	if err = s.auth(); err != nil {
		return
	}

	err = s.invoke(proto.CmdNodeInfo, proto.NodeInfoParams{Identity: identity})
	if err != nil {
		s.Close()
		return
	}

	err = s.Decodef("v", &info)
	return
}

func (s *Session) Register(service string, target string) (err error) {
	if err = s.auth(); err != nil {
		return
	}

	err = s.invoke(proto.CmdRegister, proto.RegisterParams{
		Service: service,
		Target:  target,
	})
	if err != nil {
		s.Close()
	}

	return
}

func (s *Session) Exec(identity *astral.Identity, app string, args []string, env []string) (err error) {
	if err = s.auth(); err != nil {
		return
	}

	err = s.invoke(proto.CmdExec, proto.ExecParams{
		Identity: identity,
		Exec:     app,
		Args:     args,
		Env:      env,
	})

	return err
}

func (s *Session) proto() string {
	p := strings.SplitN(s.addr, ":", 2)
	return p[0]
}

func (s *Session) auth() error {
	if err := s.conn.WriteMsg(proto.AuthParams{Token: s.token}); err != nil {
		return err
	}

	return s.conn.ReadErr()
}

func (s *Session) invoke(cmd string, params interface{}) error {
	if err := s.conn.WriteMsg(proto.Command{Cmd: cmd}); err != nil {
		return err
	}
	if err := s.conn.WriteMsg(params); err != nil {
		return err
	}
	return s.conn.ReadErr()
}
