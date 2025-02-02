package enc_test

import (
	"testing"

	"github.com/ohait/forego/enc"
	"github.com/ohait/forego/test"
)

func TestNumeric(t *testing.T) {
	c := test.Context(t)
	i := int64(1000000234567890123) // big enough to be rounded as float64
	t.Logf("i64: %d", i)
	j := enc.MustMarshalJSON(c, i)
	test.EqualsStr(t, string(j), `1000000234567890123`)
	test.NoError(t, enc.UnmarshalJSON(c, j, &i))
	test.EqualsGo(t, 1000000234567890123, i)
	var a any
	test.NoError(t, enc.UnmarshalJSON(c, j, &a))
	_ = a.(float64)
	t.Logf("a: %v", a)
}
