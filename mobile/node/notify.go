package astralmobile

import (
	"encoding/json"
	notify "github.com/cryptopunkscc/astrald/mobile/android/service/notification/go"
)

type NativeAndroidNotify interface {
	Create(channel []byte) error
	Notify(notifications []byte) error
}

var _ notify.Api = AndroidNotify{}

type AndroidNotify struct {
	native NativeAndroidNotify
}

func (an AndroidNotify) Create(channel notify.Channel) error {
	bytes, err := json.Marshal(channel)
	if err != nil {
		return err
	}
	return an.native.Create(bytes)
}

func (an AndroidNotify) Notify(notifications ...notify.Notification) error {
	bytes, err := json.Marshal(notifications[:])
	if err != nil {
		return err
	}
	return an.native.Notify(bytes)
}
