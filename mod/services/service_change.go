package services

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &ServiceChange{}

type ServiceChange struct {
	Type    ServiceChangeType
	Enabled astral.Bool
	Service Service
}

func (s ServiceChange) ObjectType() string {
	return "services.service_change"
}

func (s ServiceChange) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&s).WriteTo(w)
}

func (s *ServiceChange) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(s).ReadFrom(r)
}

func init() {
	_ = astral.DefaultBlueprints.Add(&ServiceChange{})
}

type ServiceChangeType string

const (
	ServiceChangeTypeUpdate   ServiceChangeType = "update"
	ServiceChangeTypeSnapshot                   = "snapshot"
	ServiceChangeTypeFlush                      = "flush"
)
