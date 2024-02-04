package shares

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"slices"
)

const describeServiceName = "shares.describe"

type JSONDescriptor struct {
	Type string
	Info json.RawMessage
}

type DescribeService struct {
	*Module
}

func NewDescribeService(mod *Module) *DescribeService {
	return &DescribeService{Module: mod}
}

func (srv *DescribeService) Run(ctx context.Context) error {
	err := srv.node.LocalRouter().AddRoute(describeServiceName, srv)
	if err != nil {
		return err
	}
	defer srv.node.LocalRouter().RemoveRoute(describeServiceName)

	<-ctx.Done()
	return nil
}

func (srv *DescribeService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := router.ParseQuery(query.Query())

	idParam, found := params["id"]
	if !found {
		return net.Reject()
	}

	dataID, err := data.Parse(idParam)
	if err != nil {
		srv.log.Errorv(2, "dataID parse error: %v", err)
		return net.Reject()
	}

	err = srv.Authorize(query.Caller(), dataID)
	if err != nil {
		srv.log.Errorv(2, "access to %v denied for %v (%v)", dataID, query.Caller(), err)
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		var list []JSONDescriptor

		for _, d := range srv.content.Describe(ctx, dataID, nil) {
			if !slices.Contains(srv.config.DescriptorWhitelist, d.Data.DescriptorType()) {
				continue
			}

			b, err := json.Marshal(d.Data)
			if err != nil {
				continue
			}

			list = append(list, JSONDescriptor{
				Type: d.Data.DescriptorType(),
				Info: json.RawMessage(b),
			})
		}

		json.NewEncoder(conn).Encode(list)
	})
}
