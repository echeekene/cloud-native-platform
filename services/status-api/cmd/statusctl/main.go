// Command statusctl runs the platform checks once from the command line.
// CI uses it as a smoke test after deploys: it exits 1 if anything is down.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/echeekene/cloud-native-platform/services/status-api/internal/checks"
)

func main() {
	dns := flag.String("dns", "127.0.0.1:53", "DNS server to test (host:port)")
	internal := flag.String("internal", "status.platform.internal", "name in the internal zone")
	public := flag.String("public", "ekeneeche.com", "public name to resolve recursively")
	ntp := flag.String("ntp", "127.0.0.1:123", "NTP server to test (host:port)")
	maxOffset := flag.Duration("max-offset", 100*time.Millisecond, "allowed clock offset")
	timeout := flag.Duration("timeout", 3*time.Second, "timeout per check")
	asJSON := flag.Bool("json", false, "print JSON instead of a table")
	flag.Parse()

	list := checks.Default(*dns, *internal, *public, *ntp, *maxOffset)
	results := checks.RunAll(context.Background(), *timeout, list)
	overall := checks.Overall(results)

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{"overall": overall, "results": results})
	} else {
		for _, r := range results {
			msg := r.Detail
			if r.Error != "" {
				msg = r.Error
			}
			fmt.Printf("%-24s %-9s %8.2f ms  %s\n", r.Name, r.Status, r.LatencyMS, msg)
		}
		fmt.Printf("\noverall: %s\n", overall)
	}

	if overall == checks.StatusDown {
		os.Exit(1)
	}
}
