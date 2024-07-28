package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/jxskiss/base62"
	"gorm.io/gorm"
	"io"
	"strings"
	"time"
)

const DefaultWorkerCount = 8
const infoPrefix = "node1"
const featureMux2 = "mux2"
const defaultPingTimeout = time.Second * 30

type NodeInfo nodes.NodeInfo

var _ nodes.Module = &Module{}

type Deps struct {
	Admin   admin.Module
	Dir     dir.Module
	Exonet  exonet.Module
	Keys    keys.Module
	Objects objects.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets resources.Resources
	db     *gorm.DB

	streams sig.Set[*Stream]
	conns   sig.Map[astral.Nonce, *conn]

	in chan *Frame
}

func (mod *Module) Run(ctx context.Context) error {
	go mod.frameReader(ctx)
	<-ctx.Done()
	return nil
}

func (mod *Module) Peers() (peers []id.Identity) {
	var r map[string]struct{}

	for _, s := range mod.streams.Clone() {
		if _, found := r[s.RemoteIdentity().PublicKeyHex()]; found {
			continue
		}
		r[s.RemoteIdentity().PublicKeyHex()] = struct{}{}
		peers = append(peers, s.RemoteIdentity())
	}

	return
}

func (mod *Module) ParseInfo(s string) (*nodes.NodeInfo, error) {
	trimmed := strings.TrimPrefix(s, infoPrefix)
	data, err := base62.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}

	info, err := (&InfoEncoder{mod}).Unpack(data)
	if err != nil {
		return nil, err
	}

	return (*nodes.NodeInfo)(info), nil
}

func (mod *Module) InfoString(info *nodes.NodeInfo) string {
	packed, err := (&InfoEncoder{mod}).Pack((*NodeInfo)(info))
	if err != nil {
		return ""
	}

	return infoPrefix + base62.EncodeToString(packed)
}

func (mod *Module) Resolve(ctx context.Context, identity id.Identity) ([]exonet.Endpoint, error) {
	return mod.Endpoints(identity), nil
}

func (mod *Module) frameReader(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case frame := <-mod.in:
			switch f := frame.Frame.(type) {
			case *frames.Query:
				go mod.handleQuery(frame.Source, f)
			case *frames.Response:
				mod.handleResponse(frame.Source, f)
			case *frames.Ping:
				mod.handlePing(frame.Source, f)
			case *frames.Data:
				mod.handleData(frame.Source, f)
			case *frames.Reset:
				mod.handleReset(frame.Source, f)
			case *frames.Read:
				mod.handleRead(frame.Source, f)
			default:
				mod.log.Errorv(2, "unknown frame: %v", frame.Frame)
			}
		}
	}
}

func (mod *Module) handleQuery(s *Stream, f *frames.Query) {
	conn, ok := mod.conns.Set(f.Nonce, newConn(f.Nonce))
	if !ok {
		return // ignore duplicates
	}

	conn.RemoteIdentity = s.RemoteIdentity()
	conn.Query = f.Query
	conn.stream = s
	conn.wsize = int(f.Buffer)

	var q = astral.NewQueryNonce(s.RemoteIdentity(), s.LocalIdentity(), f.Query, f.Nonce)

	w, err := mod.node.Router().RouteQuery(
		context.Background(),
		q,
		conn,
		astral.Hints{Origin: astral.OriginNetwork},
	)

	if err != nil {
		conn.swapState(stateRouting, stateClosed)
		s.Write(&frames.Response{Nonce: f.Nonce, ErrCode: frames.CodeRejected})
		return
	}

	s.Write(&frames.Response{Nonce: f.Nonce, ErrCode: frames.CodeAccepted, Buffer: uint32(conn.rsize)})
	conn.swapState(stateRouting, stateOpen)

	go func() {
		io.Copy(w, conn)
		w.Close()
	}()
}

func (mod *Module) handleResponse(s *Stream, f *frames.Response) {
	// find the connection
	conn, ok := mod.conns.Get(f.Nonce)
	if !ok {
		return
	}

	// make sure we sent the query to the identity that sent the response
	if !conn.RemoteIdentity.IsEqual(s.RemoteIdentity()) {
		return
	}

	// if rejected
	if f.ErrCode != 0 {
		if !conn.swapState(stateRouting, stateClosed) {
			return
		}
		conn.res <- false
	}

	if !conn.swapState(stateRouting, stateOpen) {
		return
	}
	conn.stream = s
	conn.wsize = int(f.Buffer)
	conn.res <- true
}

func (mod *Module) handleData(s *Stream, f *frames.Data) {
	conn, ok := mod.conns.Get(f.Nonce)
	if !ok {
		s.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	if conn.state.Load() != stateOpen {
		s.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	err := conn.pushRead(f.Payload)
	if err != nil {
		conn.Close()
		return
	}
}

func (mod *Module) handleRead(s *Stream, f *frames.Read) {
	conn, ok := mod.conns.Get(f.Nonce)
	if !ok {
		s.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	conn.growRemoteBuffer(int(f.Len))
}

func (mod *Module) handleReset(s *Stream, f *frames.Reset) {
	conn, ok := mod.conns.Get(f.Nonce)
	if !ok {
		return
	}

	conn.swapState(stateOpen, stateClosed)
}

func (mod *Module) handlePing(s *Stream, f *frames.Ping) {
	if f.Pong {
		rtt, err := s.pong(f.Nonce)
		if err != nil {
			mod.log.Errorv(1, "invalid pong nonce from %v", s.RemoteIdentity())
		} else {
			if mod.config.LogPings {
				mod.log.Logv(1, "ping with %v: %v", s.RemoteIdentity(), rtt)
			}
		}
	} else {
		s.Write(&frames.Ping{
			Nonce: f.Nonce,
			Pong:  true,
		})
	}
}

func (mod *Module) addStream(s *Stream) (err error) {
	linked := mod.isLinked(s.RemoteIdentity())

	err = mod.streams.Add(s)
	if err == nil {
		if !linked {
			mod.Objects.PushLocal(&nodes.EventLinked{NodeID: s.RemoteIdentity()})
		}

		mod.log.Infov(1, "stream with %v added", s.RemoteIdentity())
		go func() {
			for frame := range s.Read() {
				mod.in <- &Frame{
					Frame:  frame,
					Source: s,
				}
			}
			mod.log.Errorv(1, "stream with %v removed: %v", s.RemoteIdentity(), s.Err())
			mod.streams.Remove(s)
			for _, c := range mod.conns.Select(func(k astral.Nonce, v *conn) (ok bool) {
				return v.stream == s
			}) {
				c.Close()
			}

			if !mod.isLinked(s.RemoteIdentity()) {
				mod.Objects.PushLocal(&nodes.EventUnlinked{NodeID: s.RemoteIdentity()})
			}
		}()
	}

	return
}

func (mod *Module) isLinked(remoteID id.Identity) bool {
	for _, s := range mod.streams.Clone() {
		if s.RemoteIdentity().IsEqual(remoteID) {
			return true
		}
	}
	return false
}

func byRemoteID(remoteID id.Identity) func(s *Stream) bool {
	return func(s *Stream) bool {
		return s.RemoteIdentity().IsEqual(remoteID)
	}
}
