package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/echeekene/cloud-native-platform/services/status-api/internal/checks"
)

type fakeCheck struct{ r checks.Result }

func (f fakeCheck) Run(context.Context) checks.Result { return f.r }

func newTestServer(results ...checks.Result) *Server {
	list := make([]checks.Check, len(results))
	for i, r := range results {
		list[i] = fakeCheck{r}
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(list, time.Second, 100*time.Millisecond, "test", log)
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestReadyzWaitsForFirstRun(t *testing.T) {
	s := newTestServer(checks.Result{Name: "a", Kind: "dns", Status: checks.StatusUp})
	h := s.Handler()

	if rec := get(t, h, "/readyz"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("before first run: code = %d, want 503", rec.Code)
	}
	s.RunOnce(context.Background())
	if rec := get(t, h, "/readyz"); rec.Code != http.StatusOK {
		t.Errorf("after first run: code = %d, want 200", rec.Code)
	}
}

func TestStatusReturnsJSON(t *testing.T) {
	s := newTestServer(
		checks.Result{Name: "a", Kind: "dns", Status: checks.StatusUp},
		checks.Result{Name: "b", Kind: "ntp", Status: checks.StatusDegraded},
	)
	s.RunOnce(context.Background())

	rec := get(t, s.Handler(), "/api/status")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	var snap Snapshot
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatal(err)
	}
	if snap.Overall != checks.StatusDegraded || len(snap.Results) != 2 || snap.Version != "test" {
		t.Errorf("unexpected snapshot %+v", snap)
	}
}

func TestStatusIs503WhenDown(t *testing.T) {
	s := newTestServer(checks.Result{Name: "a", Kind: "dns", Status: checks.StatusDown})
	s.RunOnce(context.Background())
	if rec := get(t, s.Handler(), "/api/status"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("code = %d, want 503", rec.Code)
	}
}

func TestMetricsExposeCheckState(t *testing.T) {
	off := 12.5
	s := newTestServer(checks.Result{Name: "chrony-time", Kind: "ntp", Status: checks.StatusUp, OffsetMS: &off})
	s.RunOnce(context.Background())

	body := get(t, s.Handler(), "/metrics").Body.String()
	for _, want := range []string{
		`platform_check_up{check="chrony-time",kind="ntp"} 1`,
		`platform_ntp_offset_seconds{check="chrony-time",kind="ntp"} 0.0125`,
		`platform_checks_total{check="chrony-time",result="up"} 1`,
		`platform_build_info{version="test"} 1`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics missing %q\n%s", want, body)
		}
	}
}

func TestSecurityHeadersAndIndex(t *testing.T) {
	s := newTestServer()
	rec := get(t, s.Handler(), "/")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Platform Status") {
		t.Errorf("index not served: %d", rec.Code)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing nosniff header")
	}
	if rec := get(t, s.Handler(), "/nope"); rec.Code != http.StatusNotFound {
		t.Errorf("unknown path code = %d, want 404", rec.Code)
	}
}
