// Package config reads service settings from environment variables.
package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	ListenAddr   string
	DNSServer    string
	DNSQuery     string
	UpstreamDNS  string
	NTPServer    string
	NTPMaxOffset time.Duration
	Interval     time.Duration
	Timeout      time.Duration
}

// Load reads settings with defaults that match the Kubernetes manifests.
func Load() (Config, error) {
	c := Config{
		ListenAddr:  env("LISTEN_ADDR", ":8080"),
		DNSServer:   env("DNS_SERVER", "coredns.platform.svc.cluster.local:53"),
		DNSQuery:    env("DNS_QUERY", "status.platform.internal"),
		UpstreamDNS: env("UPSTREAM_DNS_QUERY", "ekeneeche.com"),
		NTPServer:   env("NTP_SERVER", "chrony.platform.svc.cluster.local:123"),
	}

	var err error
	if c.NTPMaxOffset, err = duration("NTP_MAX_OFFSET", "100ms"); err != nil {
		return c, err
	}
	if c.Interval, err = duration("CHECK_INTERVAL", "15s"); err != nil {
		return c, err
	}
	if c.Timeout, err = duration("CHECK_TIMEOUT", "3s"); err != nil {
		return c, err
	}
	if c.Timeout >= c.Interval {
		return c, fmt.Errorf("CHECK_TIMEOUT (%s) must be shorter than CHECK_INTERVAL (%s)", c.Timeout, c.Interval)
	}
	return c, nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func duration(key, def string) (time.Duration, error) {
	d, err := time.ParseDuration(env(key, def))
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return d, nil
}
