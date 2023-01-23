package hub

import (
	"fmt"
)

type EventPortRegistered struct {
	PortName string
}

func (e EventPortRegistered) String() string {
	return fmt.Sprintf("PortName: %s", e.PortName)
}

type EventPortReleased struct {
	PortName string
}

func (e EventPortReleased) String() string {
	return fmt.Sprintf("PortName: %s", e.PortName)
}
