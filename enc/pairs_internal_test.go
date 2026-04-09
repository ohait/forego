package enc

import "testing"

func TestPairsNativeUsesNames(t *testing.T) {
	p := Pairs{
		{Name: "foo_key", Value: Integer(3)},
		{Name: "bar_key", Value: String("baz")},
	}

	native := p.native()
	m, ok := native.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", native)
	}
	if got := m["foo_key"]; got != int64(3) {
		t.Fatalf("unexpected value for foo_key: %#v", got)
	}
	if got := m["bar_key"]; got != "baz" {
		t.Fatalf("unexpected value for bar_key: %#v", got)
	}
}
