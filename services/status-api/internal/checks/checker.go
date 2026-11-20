package checks

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status is the outcome of a single check.
type Status string

const (
	StatusUp       Status = "up"
	StatusDegraded Status = "degraded"
	StatusDown     Status = "down"
)

// Result is what one check reports back to the API.
type Result struct {
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	Target    string    `json:"target"`
	Status    Status    `json:"status"`
	LatencyMS float64   `json:"latency_ms"`
	Detail    string    `json:"detail"`
	Error     string    `json:"error,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
	OffsetMS  *float64  `json:"offset_ms,omitempty"`
	latency   time.Duration
}

// Latency returns the raw latency of the check.
func (r Result) Latency() time.Duration { return r.latency }

// Check is anything that can probe a service and report a Result.
type Check interface {
	Run(ctx context.Context) Result
}

// DNSCheck resolves a known name against our DNS server.
type DNSCheck struct {
	Name   string
	Server string
	Query  string
	Lookup func(ctx context.Context, server, name string) (DNSResult, error)
}

func (c DNSCheck) Run(ctx context.Context) Result {
	lookup := c.Lookup
	if lookup == nil {
		lookup = LookupDNS
	}
	start := time.Now()
	res, err := lookup(ctx, c.Server, c.Query)
	elapsed := time.Since(start)

	r := Result{
		Name: c.Name, Kind: "dns", Target: c.Server,
		CheckedAt: time.Now().UTC(), latency: elapsed,
		LatencyMS: ms(elapsed),
	}
	if err != nil {
		r.Status = StatusDown
		r.Error = err.Error()
		return r
	}
	r.Status = StatusUp
	r.Detail = fmt.Sprintf("%s resolved to %v", c.Query, res.Addresses)
	return r
}

// NTPCheck queries our time server and flags drift beyond MaxOffset.
type NTPCheck struct {
	Name      string
	Server    string
	MaxOffset time.Duration
	Query     func(ctx context.Context, addr string) (NTPResult, error)
}

func (c NTPCheck) Run(ctx context.Context) Result {
	query := c.Query
	if query == nil {
		query = QueryNTP
	}
	res, err := query(ctx, c.Server)

	r := Result{
		Name: c.Name, Kind: "ntp", Target: c.Server,
		CheckedAt: time.Now().UTC(),
	}
	if err != nil {
		r.Status = StatusDown
		r.Error = err.Error()
		return r
	}
	r.latency = res.RTT
	r.LatencyMS = ms(res.RTT)
	off := ms(res.Offset)
	r.OffsetMS = &off
	r.Detail = fmt.Sprintf("stratum %d, offset %.3f ms", res.Stratum, off)

	if abs(res.Offset) > c.MaxOffset {
		r.Status = StatusDegraded
		r.Error = fmt.Sprintf("offset %.3f ms exceeds limit %s", off, c.MaxOffset)
		return r
	}
	r.Status = StatusUp
	return r
}

// RunAll runs every check at the same time, each with its own timeout.
func RunAll(ctx context.Context, timeout time.Duration, list []Check) []Result {
	results := make([]Result, len(list))
	var wg sync.WaitGroup
	for i, c := range list {
		wg.Add(1)
		go func(i int, c Check) {
			defer wg.Done()
			cctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			results[i] = c.Run(cctx)
		}(i, c)
	}
	wg.Wait()
	return results
}

// Overall rolls individual results up into one status.
func Overall(results []Result) Status {
	overall := StatusUp
	for _, r := range results {
		switch r.Status {
		case StatusDown:
			return StatusDown
		case StatusDegraded:
			overall = StatusDegraded
		}
	}
	return overall
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
