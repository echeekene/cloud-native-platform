// Package metrics writes the Prometheus text exposition format by hand.
// The service only needs a handful of gauges and counters, so this keeps
// the binary free of third party dependencies.
package metrics

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

type kind string

const (
	Gauge   kind = "gauge"
	Counter kind = "counter"
)

type family struct {
	help   string
	kind   kind
	values map[string]float64 // key is the rendered label set
}

// Registry holds every metric the service exposes.
type Registry struct {
	mu       sync.Mutex
	families map[string]*family
}

func NewRegistry() *Registry {
	return &Registry{families: map[string]*family{}}
}

// Register declares a metric name with its help text and type.
func (r *Registry) Register(name, help string, k kind) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.families[name]; !ok {
		r.families[name] = &family{help: help, kind: k, values: map[string]float64{}}
	}
}

// Set stores a gauge value for the given labels.
func (r *Registry) Set(name string, labels map[string]string, v float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f := r.mustGet(name)
	f.values[renderLabels(labels)] = v
}

// Add increments a counter for the given labels.
func (r *Registry) Add(name string, labels map[string]string, v float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f := r.mustGet(name)
	f.values[renderLabels(labels)] += v
}

func (r *Registry) mustGet(name string) *family {
	f, ok := r.families[name]
	if !ok {
		panic("metrics: unregistered metric " + name)
	}
	return f
}

// WriteTo renders all metrics in a stable order.
func (r *Registry) WriteTo(w io.Writer) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	names := make([]string, 0, len(r.families))
	for n := range r.families {
		names = append(names, n)
	}
	sort.Strings(names)

	var sb strings.Builder
	for _, n := range names {
		f := r.families[n]
		fmt.Fprintf(&sb, "# HELP %s %s\n# TYPE %s %s\n", n, f.help, n, f.kind)
		keys := make([]string, 0, len(f.values))
		for k := range f.values {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&sb, "%s%s %g\n", n, k, f.values[k])
		}
	}
	n, err := io.WriteString(w, sb.String())
	return int64(n), err
}

func renderLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s=%q", k, labels[k])
	}
	return "{" + strings.Join(parts, ",") + "}"
}
