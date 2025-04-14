package ether

import "net"

type NetInterface struct {
	Flags net.Flags
	Addrs func() ([]net.Addr, error)
}

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
