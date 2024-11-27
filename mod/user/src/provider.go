package user

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/user"
	"io"
)

type Provider struct {
	*routers.PathRouter
	mod *Module
}

func NewProvider(mod *Module) *Provider {
	p := &Provider{
		mod:        mod,
		PathRouter: routers.NewPathRouter(mod.node.Identity(), false),
	}

	p.AddRouteFunc(methodNodes, p.nodes)
	p.AddRouteFunc(methodClaim, p.claim)

	return p
}

func (p *Provider) nodes(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	var args struct {
		Format string `query:"optional"`
		Names  bool   `query:"optional"`
	}
	_, err := query.ParseTo(q.Query, &args)
	if err != nil {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		nodes := p.mod.Nodes(p.mod.userID)

		switch args.Format {
		case "json":
			if args.Names {
				var names []string
				for _, n := range nodes {
					names = append(names, p.mod.Dir.DisplayName(n))
				}
				err = json.NewEncoder(conn).Encode(names)
			} else {
				err = json.NewEncoder(conn).Encode(nodes)
			}

		default:
			for _, node := range nodes {
				_, err = node.WriteTo(conn)
				if err != nil {
					return
				}
			}
		}
	})
}

func (p *Provider) claim(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if !p.mod.UserID().IsZero() {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		var contract = &user.SignedNodeContract{}
		_, err := contract.ReadFrom(conn)
		if err != nil {
			p.mod.log.Errorv(2, "claim: read contract: %v", err)
			return
		}

		if !contract.NodeID.IsEqual(p.mod.node.Identity()) {
			p.mod.log.Errorv(1, "claim: contract nodeID mismatch")
			return
		}

		if err := contract.VerifyUserSig(); err != nil {
			p.mod.log.Errorv(1, "claim: user signature invalid: %v", err)
			return
		}

		p.mod.log.Log("received claim contract from %v", contract.UserID)

		//TODO: insert a mechanism to authorize claim request

		contract.NodeSig, err = p.mod.Keys.Sign(p.mod.node.Identity(), contract.Hash())
		if err != nil {
			p.mod.log.Errorv(1, "claim: sign contract: %v", err)
			return
		}

		err = p.mod.SaveSignedNodeContract(contract)
		if err != nil {
			p.mod.log.Errorv(1, "claim: save contract: %v", err)
			return
		}

		err = p.mod.SetUserID(contract.UserID)
		if err != nil {
			p.mod.log.Errorv(1, "claim: set user: %v", err)
			return
		}

		cslq.Encode(conn, "[s]c", contract.NodeSig)
	})
}
