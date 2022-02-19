package astral

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mobile/android/node/android"
	"io"
)

type AndroidApi interface {
	Call(query string, arg string) error
	Get(query string, arg string) ([]byte, error)
	Read(query string, arg string, writer Writer) error
}

var _ android.Api = androidApi{}

type androidApi struct {
	api AndroidApi
}

func (n androidApi) Call(query string, arg interface{}) error {
	argBytes, err := json.Marshal(arg)
	if err != nil {
		return err
	}
	return n.api.Call(query, string(argBytes))
}

func (n androidApi) Get(query string, arg interface{}, result interface{}) error {
	argBytes, err := json.Marshal(arg)
	if err != nil {
		return err
	}
	resultBytes, err := n.api.Get(query, string(argBytes))
	if err != nil {
		return err
	}
	return json.Unmarshal(resultBytes, result)
}

func (n androidApi) Read(query string, arg interface{}) (reader io.ReadCloser, err error) {
	argBytes, err := json.Marshal(arg)
	if err != nil {
		return
	}
	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		_ = n.api.Read(query, string(argBytes), writer)
	}()
	return
}
