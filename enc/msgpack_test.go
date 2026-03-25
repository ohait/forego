package enc_test

import (
	"testing"
	"time"

	"github.com/ohait/forego/enc"
	"github.com/ohait/forego/test"
	"github.com/vmihailenco/msgpack/v5"
)

func TestMsgPack(t *testing.T) {
	c := test.Context(t)
	codec := &enc.MsgPack{}

	check := func(t *testing.T, nodeIn enc.Node) {
		t.Helper()
		data := codec.Encode(c, nodeIn)
		nodeOut, err := codec.Decode(c, data)
		test.NoError(t, err)
		test.EqualsGo(t, nodeIn, nodeOut)
	}

	checkLeft := func(t *testing.T, nodeIn enc.Node) {
		t.Helper()
		data := codec.Encode(c, nodeIn)
		nodeOut, err := codec.Decode(c, data)
		test.NoError(t, err)
		test.EqualsJSON(t, nodeIn, nodeOut)
	}

	t.Run("scalars", func(t *testing.T) {
		check(t, enc.Nil{})
		check(t, enc.Integer(1))
		check(t, enc.Integer(-1))
		check(t, enc.Float(3.14))
		check(t, enc.Bool(true))
		check(t, enc.String("foo"))
		check(t, enc.Bytes{1, 2, 3})
		check(t, enc.String(`\"`))
		check(t, enc.Time(time.Unix(123, 456).UTC()))
		checkLeft(t, enc.Duration(3*time.Second+250*time.Millisecond))
	})

	t.Run("digits", func(t *testing.T) {
		checkLeft(t, enc.Digits("3.14"))
		checkLeft(t, enc.Digits("3"))
		checkLeft(t, enc.Digits("18446744073709551615"))
	})

	t.Run("maps", func(t *testing.T) {
		check(t, enc.Map{})
		checkLeft(t, enc.Map{"one": enc.Float(3.14)})
		checkLeft(t, enc.Map{"one": enc.Integer(1), "nil": enc.Nil{}, "foo": enc.String("bar")})
	})

	t.Run("pairs", func(t *testing.T) {
		checkLeft(t, enc.Pairs{})
		checkLeft(t, enc.Pairs{{"b", "b", enc.Integer(1)}, {"a", "a", enc.Integer(2)}, {"", "", enc.Nil{}}})
	})

	t.Run("lists", func(t *testing.T) {
		check(t, enc.List{})
		check(t, enc.List{enc.Nil{}})
		check(t, enc.List{enc.Integer(1), enc.String("two"), enc.Bool(false)})
	})

	t.Run("deep", func(t *testing.T) {
		check(t, enc.List{enc.Map{}})
		check(t, enc.List{enc.List{}, enc.Nil{}})
		check(t, enc.Map{"l": enc.List{}})
	})
}

func TestMsgPackHelpers(t *testing.T) {
	c := test.Context(t)

	type X struct {
		S string `json:"s"`
		I int    `json:"i"`
	}

	in := X{S: "str", I: 42}
	data, err := enc.MarshalMsgPack(c, in)
	test.NoError(t, err)

	var out X
	err = enc.UnmarshalMsgPack(c, data, &out)
	test.NoError(t, err)
	test.EqualsGo(t, in, out)
}

func TestMsgPackNoData(t *testing.T) {
	c := test.Context(t)
	{
		n, err := enc.MsgPack{}.Decode(c, nil)
		test.NoError(t, err)
		test.Nil(t, n)
	}
	{
		n, err := enc.MsgPack{}.Decode(c, []byte{})
		test.NoError(t, err)
		test.Nil(t, n)
	}
}

func TestMsgPackBinary(t *testing.T) {
	c := test.Context(t)
	data, err := msgpack.Marshal([]byte{1, 2, 3})
	test.NoError(t, err)

	n, err := enc.MsgPack{}.Decode(c, data)
	test.NoError(t, err)
	test.EqualsGo(t, enc.Bytes{1, 2, 3}, n)

	var out []byte
	err = enc.Unmarshal(c, n, &out)
	test.NoError(t, err)
	test.EqualsGo(t, []byte{1, 2, 3}, out)
}
