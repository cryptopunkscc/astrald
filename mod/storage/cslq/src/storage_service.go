package cslq

import (
	"context"
	"encoding/base64"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	client "github.com/cryptopunkscc/astrald/lib/storage/cslq"
	"github.com/cryptopunkscc/astrald/mod/storage"
	api "github.com/cryptopunkscc/astrald/mod/storage/cslq"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
	"strings"
)

type StorageService struct {
	node node.Node
	mod  storage.Module
	port string
}

func NewStorageService(node node.Node, mod storage.Module) *StorageService {
	return &StorageService{node: node, mod: mod, port: api.Port}
}

func (s *StorageService) Run(ctx context.Context) (err error) {
	err = s.node.LocalRouter().AddRoute(s.port+"*", s)
	if err != nil {
		return
	}
	defer s.node.LocalRouter().RemoveRoute(s.port + "*")
	<-ctx.Done()
	return nil
}

func (s *StorageService) RouteQuery(_ context.Context, query net.Query, caller net.SecureWriteCloser, _ net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		if err := s.decode(conn, query.Caller(), query.Query()); err != nil {
			log.Println(err)
			return
		}
	})
}

func (s *StorageService) decode(
	conn io.ReadWriter,
	remoteID id.Identity,
	query string,
) (err error) {
	r := strings.NewReader(query[len(s.port):])
	d := base64.NewDecoder(client.BaseEncoding, r)
	args := cslq.NewDecoder(d)
	var cmd byte
	if err = args.Decodef("c", &cmd); err != nil {
		return
	}
	if err = s.handle(remoteID, cmd, conn, args); err != nil {
		return
	}
	return
}

func (s *StorageService) handle(
	remoteID id.Identity,
	cmd byte,
	conn io.ReadWriter,
	args *cslq.Decoder,
) (err error) {
	switch cmd {
	case api.ReadAll:
		var i data.ID
		var opts storage.OpenOpts
		if err = args.Decodef("v {q c c}", &i, &opts); err != nil {
			return
		}
		var all []byte
		if all, err = s.mod.ReadAll(i, &opts); err != nil {
			return
		}
		if err = cslq.Encode(conn, "[l]c", all); err != nil {
			return
		}
	case api.Put:
		var bytes []byte
		var opts storage.CreateOpts
		if err = args.Decodef("[l]c v", &bytes, &opts); err != nil {
			return
		}
		var dataID data.ID
		if dataID, err = s.mod.Put(bytes, &opts); err != nil {
			return
		}
		if err = cslq.Encode(conn, "v", dataID); err != nil {
			return
		}
	case api.AddOpener:
		var name string
		var port string
		var priority int
		if err = args.Decodef("[c]c [c]c l", &name, &port, &priority); err != nil {
			return
		}
		c := client.NewTargetClient(remoteID, port)
		if err = s.mod.AddOpener(name, c, priority); err != nil {
			return
		}
	case api.RemoveOpener:
		var name string
		if err = args.Decodef("[c]c", &name); err != nil {
			return
		}
		if err = s.mod.RemoveOpener(name); err != nil {
			return
		}
	case api.AddCreator:
		var name string
		var port string
		var priority int
		if err = args.Decodef("[c]c [c]c l", &name, &port, &priority); err != nil {
			return
		}
		c := client.NewTargetClient(remoteID, port)
		if err = s.mod.AddCreator(name, c, priority); err != nil {
			return
		}
	case api.RemoveCreator:
		var name string
		if err = args.Decodef("[c]c", &name); err != nil {
			return
		}
		if err = s.mod.RemoveCreator(name); err != nil {
			return
		}
	case api.AddPurger:
		var name string
		var port string
		if err = args.Decodef("[c]c [c]c", &name, &port); err != nil {
			return
		}
		c := client.NewTargetClient(remoteID, port)
		if err = s.mod.AddPurger(name, c); err != nil {
			return
		}
	case api.RemovePurger:
		var name string
		if err = args.Decodef("[c]c", &name); err != nil {
			return
		}
		if err = s.mod.RemovePurger(name); err != nil {
			return
		}
	case api.OpenerOpen:
		if err = client.NewOpenerService(s.mod, remoteID).Handle(conn, args); err != nil {
			return
		}
	case api.CreatorCreate:
		if err = client.NewCreatorService(s.mod).Handle(conn, args); err != nil {
			return
		}
	case api.PurgerPurge:
		if err = client.NewPurgerService(s.mod).Handle(conn, args); err != nil {
			return
		}
	}
	return err
}
