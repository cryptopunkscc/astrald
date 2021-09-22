package mon

import (
	"context"
	"net"
	"time"
)

// Addrs monitors the interface for new addresses
func Addrs(ctx context.Context, ifaceName string) <-chan string {
	output := make(chan string)

	go func() {
		defer close(output)

		cached := make(map[string]struct{})

		for {
			// Check if interface still exists
			iface, err := net.InterfaceByName(ifaceName)
			if err != nil {
				return
			}

			// Fetch interface addresses
			updated, err := listAddrs(iface)
			if err != nil {
				return
			}

			// Check for addresses that were removed
			for addr := range cached {
				if _, found := updated[addr]; !found {
					delete(cached, addr)
				}
			}

			// Check for addresses that were added
			for addr := range updated {
				if _, found := cached[addr]; !found {
					cached[addr] = struct{}{}
					output <- addr
				}
			}

			// Wait between polls
			select {
			case <-time.After(pollInterval):
				continue
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}

// listAddrs returns interface's addresses as a map
func listAddrs(iface *net.Interface) (map[string]struct{}, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	m := make(map[string]struct{}, len(addrs))

	for _, addr := range addrs {
		m[addr.String()] = struct{}{}
	}

	return m, nil
}
