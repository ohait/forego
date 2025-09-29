package enc_test

import (
	"math"
	"strconv"
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

func TestUint64RoundTrip(t *testing.T) {
	c := test.Context(t)
	const big = uint64(math.MaxInt64) + 1
	type S struct {
		U uint64 `json:"u"`
	}

	in := S{U: big}
	node, err := enc.Marshal(c, in)
	test.NoError(t, err)
	pairs := node.(enc.Pairs)
	v := pairs.Find("u")
	_, ok := v.(enc.Digits)
	test.Assert(t, ok)
	d := v.(enc.Digits)
	test.EqualsStr(t, strconv.FormatUint(big, 10), string(d))

	j := enc.JSON{}.Encode(c, node)
	test.EqualsStr(t, `{"u":`+strconv.FormatUint(big, 10)+`}`, string(j))

	var out S
	test.NoError(t, enc.UnmarshalJSON(c, j, &out))
	test.EqualsGo(t, in, out)
}

func TestUint64Decode(t *testing.T) {
	c := test.Context(t)
	const big = uint64(math.MaxUint64)
	node := enc.Map{"u": enc.Digits(strconv.FormatUint(big, 10))}
	var out struct {
		U uint64 `json:"u"`
	}
	test.NoError(t, enc.Unmarshal(c, node, &out))
	test.EqualsGo(t, big, out.U)
}
