package ether

import "net"

// NetInterface is a minimal abstraction over net.Interface, exposing only the fields used by broadcast logic.
// why: allows test injection without depending on real network interfaces.
type NetInterface struct {
	Flags net.Flags
	Addrs func() ([]net.Addr, error)
}

// NetInterfaces is a replaceable function that enumerates network interfaces.
// Override in tests to inject a controlled interface list.
var NetInterfaces = DefaultNetInterfaces

func DefaultNetInterfaces() (out []NetInterface, err error) {
	arr, err := net.Interfaces()
	for _, i := range arr {
		out = append(out, NetInterface{
			Flags: i.Flags,
			Addrs: i.Addrs,
		})
	}
	return
}
