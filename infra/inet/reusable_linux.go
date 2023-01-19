package inet

import (
	"golang.org/x/sys/unix"
	"net"
	"syscall"
)

func rawControl(rawConn syscall.RawConn) error {
	var err error
	// See syscall.RawConn.Control
	rawConn.Control(func(fd uintptr) {
		// Set SO_REUSEADDR
		err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		if err != nil {
			return
		}

		// Set SO_REUSEPORT
		err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		if err != nil {
			return
		}
	})
	return err
}

// See net.ListenConfig and net.Dialer's Control attribute
func control(network string, address string, rawConn syscall.RawConn) error {
	// See syscall.RawConn.Control
	if err := rawControl(rawConn); err != nil {
		return err
	}
	return nil
}

func init() {
	portConfig = net.ListenConfig{
		Control: control,
	}
	dialConfig = net.Dialer{
		Control: control,
	}
}
