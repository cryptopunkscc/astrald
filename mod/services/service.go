package services

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &Service{}

type Service struct {
	Name        astral.String8
	Identity    *astral.Identity // if its module of node then identity = NodeId, if it's of app service then identity = AppId
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

func init() {
	_ = astral.DefaultBlueprints.Add(&Service{})
}
