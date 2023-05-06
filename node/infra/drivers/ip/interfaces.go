package ip

import (
	"context"
	"net"
	"time"
)

const pollInterval = time.Second

// Interfaces monitors the system for new interfaces
func Interfaces(ctx context.Context) <-chan string {
	output := make(chan string)

	go func() {
		defer close(output)

		cached := make(map[string]struct{}, 0)

		for {
			// poll network interface names
			updated, err := listInterfaces()
			if err != nil {
				return
			}

			// Check which interfaces went down
			for name := range cached {
				if _, found := updated[name]; !found {
					delete(cached, name)
				}
			}

			// Check which interfaces went up
			for name := range updated {
				if _, found := cached[name]; !found {
					cached[name] = struct{}{}
					output <- name
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

// listInterfaces returns current network interfaces as a map
func listInterfaces() (map[string]struct{}, error) {
	names := make(map[string]struct{}, 0)

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		names[iface.Name] = struct{}{}
	}

	return names, nil
}
