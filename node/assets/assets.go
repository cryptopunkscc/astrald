package assets

import (
	"github.com/cryptopunkscc/astrald/resources"
	"gorm.io/gorm"
)

type Assets interface {
	Res() resources.Resources
	Read(name string) ([]byte, error)
	Write(name string, data []byte) error
	LoadYAML(name string, out interface{}) error
	StoreYAML(name string, in interface{}) error
	OpenDB(name string) (*gorm.DB, error)
}
