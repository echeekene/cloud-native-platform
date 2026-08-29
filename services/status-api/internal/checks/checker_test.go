package checks

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDNSCheckReportsUpAndDown(t *testing.T) {
	ok := DNSCheck{Name: "dns", Server: "x:53", Query: "a",
		Lookup: func(context.Context, string, string) (DNSResult, error) {
			return DNSResult{Addresses: []string{"10.0.0.1"}}, nil
		}}
	if r := ok.Run(context.Background()); r.Status != StatusUp {
		t.Errorf("status = %s, want up", r.Status)
	}

	bad := DNSCheck{Name: "dns", Server: "x:53", Query: "a",
		Lookup: func(context.Context, string, string) (DNSResult, error) {
			return DNSResult{}, errors.New("SERVFAIL")
		}}
	r := bad.Run(context.Background())
	if r.Status != StatusDown || r.Error == "" {
		t.Errorf("got %s with error %q, want down with an error", r.Status, r.Error)
	}
}

func TestNTPCheckFlagsDrift(t *testing.T) {
	cases := []struct {
		offset time.Duration
		want   Status
	}{
		{10 * time.Millisecond, StatusUp},
		{-10 * time.Millisecond, StatusUp},
		{150 * time.Millisecond, StatusDegraded},
		{-150 * time.Millisecond, StatusDegraded},
	}
	for _, tc := range cases {
		c := NTPCheck{Name: "ntp", Server: "x:123", MaxOffset: 100 * time.Millisecond,
			Query: func(context.Context, string) (NTPResult, error) {
				return NTPResult{Offset: tc.offset, RTT: time.Millisecond, Stratum: 3}, nil
			}}
		if got := c.Run(context.Background()).Status; got != tc.want {
			t.Errorf("offset %s: status = %s, want %s", tc.offset, got, tc.want)
		}
	}
}

func TestOverall(t *testing.T) {
	up := Result{Status: StatusUp}
	deg := Result{Status: StatusDegraded}
	down := Result{Status: StatusDown}

	if got := Overall([]Result{up, up}); got != StatusUp {
		t.Errorf("all up = %s", got)
	}
	if got := Overall([]Result{up, deg}); got != StatusDegraded {
		t.Errorf("one degraded = %s", got)
	}
	if got := Overall([]Result{deg, down, up}); got != StatusDown {
		t.Errorf("one down = %s", got)
	}
}

type slowCheck struct{}

func (slowCheck) Run(ctx context.Context) Result {
	<-ctx.Done()
	return Result{Name: "slow", Status: StatusDown, Error: ctx.Err().Error()}
}

func TestRunAllEnforcesTimeout(t *testing.T) {
	start := time.Now()
	res := RunAll(context.Background(), 50*time.Millisecond, []Check{slowCheck{}, slowCheck{}})
	if time.Since(start) > time.Second {
		t.Fatal("RunAll did not respect the timeout")
	}
	if len(res) != 2 || res[0].Status != StatusDown {
		t.Errorf("unexpected results %+v", res)
	}
}
