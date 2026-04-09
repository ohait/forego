package enc

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/ohait/forego/ctx"
)

func MustMap(n Node) Map {
	switch n := n.(type) {
	case Map:
		return n
	case Pairs:
		return n.AsMap()
	default:
		panic(fmt.Sprintf("not a map: %T", n))
	}
}

func AsMap(c ctx.C, n Node) (Map, error) {
	switch n := n.(type) {
	case Map:
		return n, nil
	case Pairs:
		return n.AsMap(), nil
	default:
		return nil, ctx.NewErrorf(c, "not a map: %T", n)
	}
}

// NOTE(oha): not sure if we really need this, left around for now might either add test or remove later
type Pipe struct {
	remoteClose chan struct{}
	Send        chan<- Node
	Recv        <-chan Node
}

var _ ReadWriteCloser = Pipe{}

func NewPipe(buf int) (Pipe, Pipe) {
	send := make(chan Node, buf)
	recv := make(chan Node, buf)
	return Pipe{
			make(chan struct{}),
			send, recv,
		},
		Pipe{
			make(chan struct{}),
			recv, send,
		}
}

func (this Pipe) Read(c ctx.C) (Node, error) {
	select {
	case n, ok := <-this.Recv:
		if !ok {
			return nil, io.EOF
		}
		return n, nil
	case <-c.Done():
		return nil, c.Err()
	}
}

func (this Pipe) Write(c ctx.C, n Node) (err error) {
	defer func() {
		if recover() != nil { // ugly, but simplier this way, allow to write on closed and get an error not a panic
			err = io.ErrClosedPipe
		}
	}()
	select {
	case this.Send <- n:
		return nil
	case <-c.Done():
		return c.Err()
	}
}

func (this Pipe) Close(c ctx.C) error {
	defer func() {
		_ = recover() // ugly, but simplier this way, allowing multiple closes with no side effects
	}()
	close(this.Send)
	return nil
}

type Tag struct {
	Name      string
	OmitEmpty bool
	Skip      bool
}

func parseTag(tag reflect.StructField) (out Tag, err error) {
	out.Name = tag.Name

	names := []string{tag.Tag.Get("name")}
	skip := false
	sawSkip := false
	for _, key := range []string{"json", "yaml", "msgpack"} {
		raw := tag.Tag.Get(key)
		if raw == "-" {
			skip = true
			sawSkip = true
			continue
		}
		name, extra, _ := strings.Cut(raw, ",")
		if extra != "" {
			for _, opt := range strings.Split(extra, ",") {
				switch opt {
				case "omitempty", "omitzero", "inline":
				case "":
				default:
					return out, fmt.Errorf("invalid %s tag on %v: %q", key, tag, opt)
				}
			}
			if key == "json" {
				for _, opt := range strings.Split(extra, ",") {
					switch opt {
					case "omitempty", "omitzero":
						out.OmitEmpty = true
					}
				}
			}
		}
		names = append(names, name)
	}

	for _, name := range names {
		if name == "" {
			continue
		}
		if sawSkip {
			return out, fmt.Errorf("conflicting field config on %v: skip and name %q", tag, name)
		}
		if out.Name == tag.Name {
			out.Name = name
			continue
		}
		if out.Name != name {
			return out, fmt.Errorf("conflicting field names on %v: %q != %q", tag, out.Name, name)
		}
	}
	out.Skip = skip
	return out, nil
}
