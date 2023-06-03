package assets

import (
	"gorm.io/gorm"
)

type Store interface {
	Read(name string) ([]byte, error)
	Write(name string, data []byte) error
	LoadYAML(name string, out interface{}) error
	StoreYAML(name string, in interface{}) error
	OpenDB(name string) (*gorm.DB, error)
	KeyStore() (KeyStore, error)
}
