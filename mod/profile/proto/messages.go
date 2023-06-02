package proto

import "time"

type Endpoint struct {
	Network   string    `json:"network,omitempty"`
	Address   string    `json:"address,omitempty"`
	Public    bool      `json:"public,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Profile struct {
	Alias     string     `json:"alias,omitempty"`
	Endpoints []Endpoint `json:"endpoints,omitempty"`
}
