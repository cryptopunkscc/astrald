package astralmobile

import (
	"encoding/json"
	content "github.com/cryptopunkscc/astrald/mobile/android/service/content/go"
	"io"
)

type NativeAndroidContentResolver interface {
	Info(uri string) (files []byte, err error)
	Read(uri string, writer Writer) error
}

type Writer io.Writer

var _ content.Api = AndroidContentResolver{}

type AndroidContentResolver struct {
	native NativeAndroidContentResolver
}

func (a AndroidContentResolver) Reader(uri string) (io.ReadCloser, error) {
	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		err := a.native.Read(uri, writer)
		if err != nil {
			return
		}
	}()
	return reader, nil
}

func (a AndroidContentResolver) Info(uri string) (files content.Info, err error) {
	bytes, err := a.native.Info(uri)
	if err != nil {
		return
	}
	err = json.Unmarshal(bytes, &files)
	return
}
