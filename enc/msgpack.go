package enc

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ohait/forego/ctx"
	"github.com/vmihailenco/msgpack/v5"
)

func MustMarshalMsgPack(c ctx.C, from any) []byte {
	n, err := Marshal(c, from)
	if err != nil {
		panic(err)
	}
	return MsgPack{}.Encode(c, n)
}

func MarshalMsgPack(c ctx.C, from any) ([]byte, error) {
	n, err := Marshal(c, from)
	if err != nil {
		return nil, err
	}
	return MsgPack{}.Encode(c, n), nil
}

func UnmarshalMsgPack(c ctx.C, data []byte, into any) error {
	n, err := MsgPack{}.Decode(c, data)
	if err != nil {
		return err
	}
	return Unmarshal(c, n, into)
}

type MsgPack struct{}

var _ Codec = MsgPack{}

func (MsgPack) Encode(c ctx.C, n Node) []byte {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	err := encodeMsgPackNode(enc, n)
	if err != nil {
		panic(ctx.WrapError(c, err))
	}
	return buf.Bytes()
}

func (MsgPack) Decode(c ctx.C, data []byte) (Node, error) {
	if len(data) == 0 {
		return nil, nil
	}
	dec := msgpack.NewDecoder(bytes.NewReader(data))
	v, err := dec.DecodeInterface()
	if err != nil {
		return nil, ctx.NewErrorf(ctx.WithTag(c, "len", len(data)), "%w", err)
	}
	n, err := fromMsgPackNative(v)
	if err != nil {
		return nil, ctx.NewErrorf(ctx.WithTag(c, "len", len(data)), "%w", err)
	}
	return n, nil
}

func encodeMsgPackNode(enc *msgpack.Encoder, n Node) error {
	switch n := n.(type) {
	case nil:
		return enc.EncodeNil()
	case Nil:
		return enc.EncodeNil()
	case String:
		return enc.EncodeString(string(n))
	case Bytes:
		return enc.EncodeBytes([]byte(n))
	case Bool:
		return enc.EncodeBool(bool(n))
	case Integer:
		return enc.EncodeInt(int64(n))
	case Float:
		return enc.EncodeFloat64(float64(n))
	case Digits:
		if n.IsFloat() {
			f, err := n.Float64()
			if err != nil {
				return err
			}
			return enc.EncodeFloat64(f)
		}
		i, err := n.Int64()
		if err == nil {
			return enc.EncodeInt(i)
		}
		u, err := n.Uint64()
		if err == nil {
			return enc.EncodeUint(u)
		}
		return err
	case List:
		err := enc.EncodeArrayLen(len(n))
		if err != nil {
			return err
		}
		for _, item := range n {
			err = encodeMsgPackNode(enc, item)
			if err != nil {
				return err
			}
		}
		return nil
	case Map:
		err := enc.EncodeMapLen(len(n))
		if err != nil {
			return err
		}
		for k, v := range n {
			err = enc.EncodeString(k)
			if err != nil {
				return err
			}
			err = encodeMsgPackNode(enc, v)
			if err != nil {
				return err
			}
		}
		return nil
	case Pairs:
		err := enc.EncodeMapLen(len(n))
		if err != nil {
			return err
		}
		for _, p := range n {
			err = enc.EncodeString(p.Name)
			if err != nil {
				return err
			}
			err = encodeMsgPackNode(enc, p.Value)
			if err != nil {
				return err
			}
		}
		return nil
	case Time:
		return enc.EncodeTime(time.Time(n))
	case Duration:
		return enc.EncodeString(n.String())
	default:
		return fmt.Errorf("msgpack encode unsupported node %T", n)
	}
}

func fromMsgPackNative(in any) (Node, error) {
	switch in := in.(type) {
	case nil:
		return Nil{}, nil
	case bool:
		return Bool(in), nil
	case string:
		return String(in), nil
	case []byte:
		out := make(Bytes, len(in))
		copy(out, in)
		return out, nil
	case int8:
		return Integer(in), nil
	case int16:
		return Integer(in), nil
	case int32:
		return Integer(in), nil
	case int64:
		return Integer(in), nil
	case uint8:
		return Integer(in), nil
	case uint16:
		return Integer(in), nil
	case uint32:
		return Integer(in), nil
	case uint64:
		if in <= math.MaxInt64 {
			return Integer(in), nil
		}
		return Digits(strconv.FormatUint(in, 10)), nil
	case float32:
		return Float(in), nil
	case float64:
		return Float(in), nil
	case time.Time:
		return Time(in.UTC()), nil
	case []any:
		out := make(List, 0, len(in))
		for _, v := range in {
			n, err := fromMsgPackNative(v)
			if err != nil {
				return nil, err
			}
			out = append(out, n)
		}
		return out, nil
	case map[string]any:
		out := Map{}
		for k, v := range in {
			n, err := fromMsgPackNative(v)
			if err != nil {
				return nil, err
			}
			out[k] = n
		}
		return out, nil
	case map[any]any:
		out := Map{}
		for k, v := range in {
			ks, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("msgpack map key %T is not supported; only string keys are supported", k)
			}
			n, err := fromMsgPackNative(v)
			if err != nil {
				return nil, err
			}
			out[ks] = n
		}
		return out, nil
	default:
		return nil, fmt.Errorf("msgpack value %T is not supported", in)
	}
}
