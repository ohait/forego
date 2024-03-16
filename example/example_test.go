package example_test

import (
	"testing"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/test"
)

func lib(c ctx.C) {
	log.Debugf(c, "foobar")
}

func TestAll(t *testing.T) {
	c := test.Context(t)
	t.Logf("before")
	lib(c)
	t.Logf("after")
	//test.Assert(t, false)
}
