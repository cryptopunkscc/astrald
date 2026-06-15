package ip

import "net"

// InterfaceAddrs is an indirection over net.InterfaceAddrs to allow substitution in tests.
var InterfaceAddrs = net.InterfaceAddrs
