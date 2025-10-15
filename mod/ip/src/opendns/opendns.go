package opendns

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"
)

const myIPDomain = "myip.opendns.com"
const lookupTimeout = 5 * time.Second
const dialTimeout = 5 * time.Second

// List of OpenDNS resolvers
var resolvers = []string{
	"resolver1.opendns.com:53",
	"resolver2.opendns.com:53",
	"resolver3.opendns.com:53",
	"resolver4.opendns.com:53",
}

func LookupMyIP() (string, error) {
	// Randomly select a resolver
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	dnsServer := resolvers[r.Intn(len(resolvers))]

	// Create a DNS resolver
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: dialTimeout,
			}
			return d.DialContext(ctx, network, dnsServer)
		},
	}

	// Perform DNS lookup
	ctx, cancel := context.WithTimeout(context.Background(), lookupTimeout)
	defer cancel()

	ips, err := resolver.LookupHost(ctx, myIPDomain)
	if err != nil {
		return "", fmt.Errorf("DNS lookup failed: %w", err)
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("DNS lookup failed: no records found")
	}

	return ips[0], nil
}
