package prom_test

import (
	"bytes"
	"testing"

	"github.com/ohait/forego/test"
	"github.com/ohait/forego/utils/prom"
)

func TestCounter(t *testing.T) {

	m := prom.Counter{
		Labels: []string{"path", "op"},
	}
	m.Observe(3.14, "/foo", "read")
	m.Observe(0.5, "/foo", "write")
	m.Observe(1.0, "/foo", "write")
	m.Observe(2, "/bar", "read")

	w := &bytes.Buffer{}
	if err := m.Print(t.Name(), w); err != nil {
		t.Fatal(err)
	}
	t.Logf("full: \n%s", w.String())
	test.Contains(t, w.String(), "/foo")
	test.Contains(t, w.String(), "1.5") // 0.5 + 1.0

}
