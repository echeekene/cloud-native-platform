package checks

import (
	"context"
	"net"
	"testing"
	"time"
)

// startFakeNTP runs a tiny SNTP server whose clock is shifted by skew.
func startFakeNTP(t *testing.T, skew time.Duration, stratum byte) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })

	go func() {
		buf := make([]byte, 48)
		for {
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			if n < 48 {
				continue
			}
			now := time.Now().Add(skew)
			resp := make([]byte, 48)
			resp[0] = 0x24 // version 4, mode 4 (server)
			resp[1] = stratum
			copy(resp[24:32], buf[40:48]) // originate = client transmit
			putTimestamp(resp[32:40], now)
			putTimestamp(resp[40:48], now)
			_, _ = pc.WriteTo(resp, addr)
		}
	}()
	return pc.LocalAddr().String()
}

func TestQueryNTPMeasuresOffset(t *testing.T) {
	addr := startFakeNTP(t, 250*time.Millisecond, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := QueryNTP(ctx, addr)
	if err != nil {
		t.Fatalf("QueryNTP: %v", err)
	}
	if res.Stratum != 2 {
		t.Errorf("stratum = %d, want 2", res.Stratum)
	}
	if res.Offset < 200*time.Millisecond || res.Offset > 300*time.Millisecond {
		t.Errorf("offset = %s, want about 250ms", res.Offset)
	}
}

func TestQueryNTPRejectsKissOfDeath(t *testing.T) {
	addr := startFakeNTP(t, 0, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := QueryNTP(ctx, addr); err == nil {
		t.Fatal("expected error for stratum 0 response")
	}
}

func TestQueryNTPTimesOut(t *testing.T) {
	// A UDP socket that never answers.
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if _, err := QueryNTP(ctx, pc.LocalAddr().String()); err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestTimestampRoundTrip(t *testing.T) {
	want := time.Date(2026, 11, 20, 12, 0, 0, 500_000_000, time.UTC)
	b := make([]byte, 8)
	putTimestamp(b, want)
	got := readTimestamp(b)
	if diff := got.Sub(want); diff < -time.Microsecond || diff > time.Microsecond {
		t.Errorf("round trip drift %s", diff)
	}
}

func TestParseRejectsShortPacket(t *testing.T) {
	if _, err := parseNTPResponse(make([]byte, 10), make([]byte, 8), time.Now(), time.Now()); err == nil {
		t.Fatal("expected error for short packet")
	}
}
