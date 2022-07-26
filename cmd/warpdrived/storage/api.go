package storage

import (
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"io"
	"os"
)

type Offer interface {
	Save(offer warpdrive.Offer)
	Get() warpdrive.Offers
}

type Peer interface {
	Save(peers []warpdrive.Peer)
	Get() warpdrive.Peers
	List() []warpdrive.Peer
}

type File interface {
	IsExist(err error) bool
	MkDir(path string, perm os.FileMode) error
	FileWriter(path string, perm os.FileMode) (io.WriteCloser, error)
}

// FileResolver provides file reader for uri.
// Required for platforms where direct access to the file system is restricted.
type FileResolver interface {
	Reader(uri string) (io.ReadCloser, error)
	Info(uri string) (files []warpdrive.Info, err error)
}
