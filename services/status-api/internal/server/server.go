package server

import (
	"context"
	"embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/echeekene/cloud-native-platform/services/status-api/internal/checks"
	"github.com/echeekene/cloud-native-platform/services/status-api/internal/metrics"
)

//go:embed web/index.html
var webFS embed.FS

// Snapshot is the latest set of results served by the API.
type Snapshot struct {
	Overall   checks.Status   `json:"overall"`
	Results   []checks.Result `json:"results"`
	UpdatedAt time.Time       `json:"updated_at"`
	Version   string          `json:"version"`
}

// Server runs the checks on a timer and serves the latest snapshot.
type Server struct {
	Checks   []checks.Check
	Interval time.Duration
	Timeout  time.Duration
	Version  string
	Log      *slog.Logger

	reg  *metrics.Registry
	mu   sync.RWMutex
	snap *Snapshot
}

func New(list []checks.Check, interval, timeout time.Duration, version string, log *slog.Logger) *Server {
	reg := metrics.NewRegistry()
	reg.Register("platform_check_up", "1 if the check is up, 0.5 if degraded, 0 if down.", metrics.Gauge)
	reg.Register("platform_check_latency_seconds", "Latency of the most recent check.", metrics.Gauge)
	reg.Register("platform_ntp_offset_seconds", "Clock offset reported by the NTP check.", metrics.Gauge)
	reg.Register("platform_checks_total", "Checks run, by result.", metrics.Counter)
	reg.Register("platform_build_info", "Build information.", metrics.Gauge)
	reg.Set("platform_build_info", map[string]string{"version": version}, 1)

	return &Server{Checks: list, Interval: interval, Timeout: timeout, Version: version, Log: log, reg: reg}
}

// RunOnce executes every check and stores the snapshot.
func (s *Server) RunOnce(ctx context.Context) *Snapshot {
	results := checks.RunAll(ctx, s.Timeout, s.Checks)
	snap := &Snapshot{
		Overall:   checks.Overall(results),
		Results:   results,
		UpdatedAt: time.Now().UTC(),
		Version:   s.Version,
	}

	for _, r := range results {
		labels := map[string]string{"check": r.Name, "kind": r.Kind}
		s.reg.Set("platform_check_up", labels, upValue(r.Status))
		s.reg.Set("platform_check_latency_seconds", labels, r.Latency().Seconds())
		if r.OffsetMS != nil {
			s.reg.Set("platform_ntp_offset_seconds", labels, *r.OffsetMS/1000)
		}
		s.reg.Add("platform_checks_total", map[string]string{"check": r.Name, "result": string(r.Status)}, 1)

		if r.Status != checks.StatusUp {
			s.Log.Warn("check not healthy", "check", r.Name, "status", r.Status, "error", r.Error)
		}
	}

	s.mu.Lock()
	s.snap = snap
	s.mu.Unlock()
	return snap
}

// Loop runs checks every Interval until ctx is cancelled.
func (s *Server) Loop(ctx context.Context) {
	s.RunOnce(ctx)
	t := time.NewTicker(s.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.RunOnce(ctx)
		}
	}
}

func (s *Server) latest() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snap
}

// Handler wires up every route.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("GET /metrics", s.handleMetrics)
	mux.HandleFunc("GET /{$}", s.handleIndex)
	return securityHeaders(mux)
}

// healthz only says the process is alive.
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// readyz waits until the first round of checks has finished.
func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	if s.latest() == nil {
		http.Error(w, "warming up", http.StatusServiceUnavailable)
		return
	}
	_, _ = w.Write([]byte("ready\n"))
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	snap := s.latest()
	if snap == nil {
		http.Error(w, `{"error":"no results yet"}`, http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if snap.Overall == checks.StatusDown {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(snap)
}

func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = s.reg.WriteTo(w)
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	page, err := webFS.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, "page missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'")
		next.ServeHTTP(w, r)
	})
}

func upValue(s checks.Status) float64 {
	switch s {
	case checks.StatusUp:
		return 1
	case checks.StatusDegraded:
		return 0.5
	default:
		return 0
	}
}
