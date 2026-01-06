package services

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &Service{}

type Service struct {
	Name        astral.String8   // mod.nat
	Identity    *astral.Identity // if its module of node then idenity = NodeId, if it's service accessible beacuase of app then identity = AppId
	Composition *astral.Bundle
}

func (s Service) ObjectType() string {
	return "services.service"
}

func (s Service) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&s).WriteTo(w)
}

func (s *Service) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(s).ReadFrom(r)
}
