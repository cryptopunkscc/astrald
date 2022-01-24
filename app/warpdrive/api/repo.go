package api

import (
	"io"
	"os"
)

type Repository interface {
	Incoming() OffersRepo
	Outgoing() OffersRepo
	Peers() PeersRepo
}

type OffersRepo interface {
	Save(offer *Offer)
	List() Offers
}

type PeersRepo interface {
	Save(peers []Peer)
	List() []Peer
}

type Storage interface {
	IsExist(err error) bool
	MkDir(path string, perm os.FileMode) error
	FileWriter(path string, perm os.FileMode) (io.WriteCloser, error)
}

// Resolver provides file reader for uri.
// Required for platforms where direct access to the file system is restricted.
type Resolver interface {
	File(uri string) (io.ReadCloser, error)
	Info(uri string) (files []Info, err error)
}
