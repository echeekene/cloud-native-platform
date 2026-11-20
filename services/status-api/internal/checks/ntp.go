package checks

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

// ntpEpochOffset is the number of seconds between the NTP epoch (1900)
// and the Unix epoch (1970).
const ntpEpochOffset = 2208988800

// NTPResult holds what a single SNTP query told us about a time server.
type NTPResult struct {
	Offset  time.Duration // how far the local clock is from the server
	RTT     time.Duration // round trip delay of the query
	Stratum uint8         // distance from the reference clock
}

// QueryNTP sends one SNTP (RFC 4330) request to addr and computes the
// clock offset and round trip delay from the four packet timestamps.
func QueryNTP(ctx context.Context, addr string) (NTPResult, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "udp", addr)
	if err != nil {
		return NTPResult{}, fmt.Errorf("dial ntp: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	req := make([]byte, 48)
	req[0] = 0x23 // LI = 0, Version = 4, Mode = 3 (client)

	t1 := time.Now()
	putTimestamp(req[40:], t1) // transmit timestamp
	if _, err := conn.Write(req); err != nil {
		return NTPResult{}, fmt.Errorf("send ntp request: %w", err)
	}

	resp := make([]byte, 48)
	n, err := conn.Read(resp)
	t4 := time.Now()
	if err != nil {
		return NTPResult{}, fmt.Errorf("read ntp response: %w", err)
	}

	return parseNTPResponse(resp[:n], req[40:48], t1, t4)
}

func parseNTPResponse(resp, sentTransmit []byte, t1, t4 time.Time) (NTPResult, error) {
	if len(resp) < 48 {
		return NTPResult{}, errors.New("ntp response too short")
	}
	if mode := resp[0] & 0x07; mode != 4 {
		return NTPResult{}, fmt.Errorf("unexpected ntp mode %d", mode)
	}
	stratum := resp[1]
	if stratum == 0 {
		return NTPResult{}, errors.New("ntp server sent kiss of death (stratum 0)")
	}
	// The server must echo our transmit timestamp as its originate timestamp.
	for i := 0; i < 8; i++ {
		if resp[24+i] != sentTransmit[i] {
			return NTPResult{}, errors.New("ntp originate timestamp mismatch")
		}
	}

	t2 := readTimestamp(resp[32:40]) // server receive
	t3 := readTimestamp(resp[40:48]) // server transmit

	offset := (t2.Sub(t1) + t3.Sub(t4)) / 2
	rtt := t4.Sub(t1) - t3.Sub(t2)
	if rtt < 0 {
		rtt = 0
	}
	return NTPResult{Offset: offset, RTT: rtt, Stratum: stratum}, nil
}

func putTimestamp(b []byte, t time.Time) {
	secs := uint64(t.Unix()) + ntpEpochOffset
	frac := (uint64(t.Nanosecond()) << 32) / 1e9
	binary.BigEndian.PutUint32(b[0:4], uint32(secs))
	binary.BigEndian.PutUint32(b[4:8], uint32(frac))
}

func readTimestamp(b []byte) time.Time {
	secs := int64(binary.BigEndian.Uint32(b[0:4])) - ntpEpochOffset
	frac := uint64(binary.BigEndian.Uint32(b[4:8]))
	nanos := int64((frac * 1e9) >> 32)
	return time.Unix(secs, nanos)
}
