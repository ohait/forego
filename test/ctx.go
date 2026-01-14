package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/enc"
)

type test string

func T(c ctx.C) *testing.T {
	if t, ok := c.Value(test("forego.t")).(*testing.T); ok {
		return t
	}
	panic("no *testing.T in context")
}

func Context(t testing.TB) ctx.C {
	t.Helper()
	c := context.Background()
	c = context.WithValue(c, test("forego.t"), t)
	c = ctx.WithTag(c, "test", t.Name())
	c = log.WithLoggerAndHelper(c, func(ln log.Line) {
		if !isTerminal { // TODO(oha) allow for an env variable to override
			fmt.Println(ln.JSON())
		} else {
			t.Helper()
			msg := ln.Message
			if ln.Tags["error"] != nil {
				var e map[string]any
				enc.UnmarshalJSON(c, []byte(ln.Tags["error"]), &e)
				if e != nil && e["stack"] != nil {
					switch st := e["stack"].(type) {
					case []any:
						for _, frame := range st {
							f := fmt.Sprint(frame)
							if strings.Contains(f, "/pkg/mod/") {
								break
							}
							msg += "        " + f + "\n"
						}
					}
				}
			}
			t.Logf("%s: %s", ln.Level, msg)
		}
	}, t.Helper)
	switch t := t.(type) {
	case *testing.T:
		d, ok := t.Deadline()
		if ok {
			c, cf := context.WithDeadline(c, d)
			t.Cleanup(cf)
			return c
		} else {
			c, cf := context.WithCancel(c)
			t.Cleanup(cf)
			return c
		}
	default:
		return c
	}
}
