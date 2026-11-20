package checks

import (
	"time"
)

// Default builds the checks the platform runs: an internal zone lookup,
// a recursive lookup through our resolver, and an NTP drift check.
func Default(dnsServer, internalName, publicName, ntpServer string, maxOffset time.Duration) []Check {
	return []Check{
		DNSCheck{Name: "coredns-internal-zone", Server: dnsServer, Query: internalName},
		DNSCheck{Name: "coredns-recursive", Server: dnsServer, Query: publicName},
		NTPCheck{Name: "chrony-time", Server: ntpServer, MaxOffset: maxOffset},
	}
}
