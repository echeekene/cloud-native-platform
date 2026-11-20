// Command status-api checks the platform's DNS and NTP services on a timer
// and serves the results as a status page, a JSON API and Prometheus metrics.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/echeekene/cloud-native-platform/services/status-api/internal/checks"
	"github.com/echeekene/cloud-native-platform/services/status-api/internal/config"
	"github.com/echeekene/cloud-native-platform/services/status-api/internal/server"
)

// version is set at build time with -ldflags "-X main.version=..."
var version = "dev"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Error("bad config", "error", err)
		os.Exit(1)
	}

	list := checks.Default(cfg.DNSServer, cfg.DNSQuery, cfg.UpstreamDNS, cfg.NTPServer, cfg.NTPMaxOffset)
	srv := server.New(list, cfg.Interval, cfg.Timeout, version, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go srv.Loop(ctx)

	httpSrv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("status api listening", "addr", cfg.ListenAddr, "version", version)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}
