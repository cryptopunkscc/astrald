package bluez

import (
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const dest = "org.bluez"
const UUID_SPP = "00001101-0000-1000-8000-00805f9b34fb"

type Bluez struct {
	bus *dbus.Conn
}

type ProfileOptions struct {
}

func New() (*Bluez, error) {
	sysBus, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	return &Bluez{
		bus: sysBus,
	}, err
}

func (bluez *Bluez) bluez() dbus.BusObject {
	return bluez.bus.Object(dest, "/org/bluez")
}

func (bluez *Bluez) RegisterProfile(uuid string) error {
	profileObject := dbus.ObjectPath("/astral/test")

	call := bluez.bluez().Call(
		"org.bluez.ProfileManager1.RegisterProfile",
		0,
		profileObject,
		uuid,
		map[string]dbus.Variant{},
	)

	return call.Err
}

func (bluez *Bluez) Adapters() ([]*Adapter, error) {
	var b = bluez.bus.Object(dest, "/org/bluez")
	var list = make([]*Adapter, 0)

	node, err := introspect.Call(b)
	if err != nil {
		return nil, err
	}

	for _, n := range node.Children {
		list = append(list, &Adapter{name: n.Name})
	}

	return list, nil
}
