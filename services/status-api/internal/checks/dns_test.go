package checks

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"
)

// startFakeDNS answers every A query with 10.0.0.53 and every other
// question type with an empty answer, which is enough for LookupHost.
func startFakeDNS(t *testing.T) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })

	go func() {
		buf := make([]byte, 512)
		for {
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			if n < 12 {
				continue
			}
			_, _ = pc.WriteTo(answer(buf[:n]), addr)
		}
	}()
	return pc.LocalAddr().String()
}

func answer(q []byte) []byte {
	// Find the end of the question name.
	i := 12
	for i < len(q) && q[i] != 0 {
		i += int(q[i]) + 1
	}
	qEnd := i + 5 // zero byte + QTYPE + QCLASS
	qtype := binary.BigEndian.Uint16(q[i+1 : i+3])

	resp := make([]byte, 0, 512)
	resp = append(resp, q[0], q[1]) // same ID
	resp = append(resp, 0x81, 0x80) // response, recursion available, no error
	resp = append(resp, 0, 1)       // QDCOUNT
	if qtype == 1 {
		resp = append(resp, 0, 1) // ANCOUNT
	} else {
		resp = append(resp, 0, 0)
	}
	resp = append(resp, 0, 0, 0, 0) // NSCOUNT, ARCOUNT
	resp = append(resp, q[12:qEnd]...)
	if qtype == 1 {
		resp = append(resp, 0xc0, 0x0c) // pointer to the question name
		resp = append(resp, 0, 1, 0, 1) // TYPE A, CLASS IN
		resp = append(resp, 0, 0, 0, 30)
		resp = append(resp, 0, 4, 10, 0, 0, 53)
	}
	return resp
}

func TestLookupDNSUsesGivenServer(t *testing.T) {
	addr := startFakeDNS(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := LookupDNS(ctx, addr, "status.platform.internal")
	if err != nil {
		t.Fatalf("LookupDNS: %v", err)
	}
	if len(res.Addresses) == 0 || res.Addresses[0] != "10.0.0.53" {
		t.Errorf("addresses = %v, want [10.0.0.53]", res.Addresses)
	}
}

func TestLookupDNSFailsWhenServerDown(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := pc.LocalAddr().String()
	pc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if _, err := LookupDNS(ctx, addr, "status.platform.internal"); err == nil {
		t.Fatal("expected error from a closed server")
	}
}
