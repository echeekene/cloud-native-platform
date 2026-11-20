package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ListenAddr != ":8080" || c.Interval != 15*time.Second || c.NTPMaxOffset != 100*time.Millisecond {
		t.Errorf("unexpected defaults %+v", c)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "30s")
	t.Setenv("DNS_SERVER", "10.0.0.10:53")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.Interval != 30*time.Second || c.DNSServer != "10.0.0.10:53" {
		t.Errorf("overrides not applied %+v", c)
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "soon")
	if _, err := Load(); err == nil {
		t.Error("expected error for bad duration")
	}
}

func TestTimeoutMustBeShorterThanInterval(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "2s")
	t.Setenv("CHECK_TIMEOUT", "5s")
	if _, err := Load(); err == nil {
		t.Error("expected error when timeout >= interval")
	}
}
