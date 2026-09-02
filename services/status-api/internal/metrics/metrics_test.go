package metrics

import (
	"strings"
	"testing"
)

func TestWriteToFormat(t *testing.T) {
	r := NewRegistry()
	r.Register("demo_total", "A demo counter.", Counter)
	r.Add("demo_total", map[string]string{"b": "2", "a": "1"}, 1)
	r.Add("demo_total", map[string]string{"a": "1", "b": "2"}, 2)

	var sb strings.Builder
	if _, err := r.WriteTo(&sb); err != nil {
		t.Fatal(err)
	}
	want := "# HELP demo_total A demo counter.\n# TYPE demo_total counter\ndemo_total{a=\"1\",b=\"2\"} 3\n"
	if sb.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", sb.String(), want)
	}
}

func TestUnregisteredMetricPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
	}()
	NewRegistry().Set("missing", nil, 1)
}
