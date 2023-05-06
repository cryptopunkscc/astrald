package bluez

import "github.com/godbus/dbus/v5"

type Adapter struct {
	name string
}

func NewAdapter(name string) *Adapter {
	return &Adapter{name: name}
}

func (a Adapter) Name() string {
	return a.name
}

func (a Adapter) Address() (string, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return "", err
	}

	v, err := bus.Object(dest, a.objectPath()).GetProperty("org.bluez.Adapter1.Address")
	if err != nil {
		return "", err
	}

	var addr string
	if err := v.Store(&addr); err != nil {
		return "", err
	}

	return addr, nil
}

func (a Adapter) objectPath() dbus.ObjectPath {
	return dbus.ObjectPath("/org/bluez/" + a.name)
}
