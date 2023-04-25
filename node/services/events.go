package services

import (
	"fmt"
)

type EventServiceRegistered struct {
	Name string
}

func (e EventServiceRegistered) String() string {
	return fmt.Sprintf("name=%s", e.Name)
}

type EventServiceReleased struct {
	Name string
}

func (e EventServiceReleased) String() string {
	return fmt.Sprintf("name=%s", e.Name)
}
