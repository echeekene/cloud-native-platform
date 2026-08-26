package checks

import (
	"context"
	"fmt"
	"net"
)

// DNSResult holds the answers a DNS server returned for one lookup.
type DNSResult struct {
	Addresses []string
}

// LookupDNS resolves name against one specific DNS server (host:port)
// instead of the system resolver, so we test the server we actually run.
func LookupDNS(ctx context.Context, server, name string) (DNSResult, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, server)
		},
	}

	addrs, err := r.LookupHost(ctx, name)
	if err != nil {
		return DNSResult{}, fmt.Errorf("lookup %s via %s: %w", name, server, err)
	}
	if len(addrs) == 0 {
		return DNSResult{}, fmt.Errorf("lookup %s via %s: no records", name, server)
	}
	return DNSResult{Addresses: addrs}, nil
}
