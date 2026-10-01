package ctx_test

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/test"
)

func TestError(t *testing.T) {
	c := ctx.TODO()
	var err error
	err = io.EOF
	t.Logf("err: %T %v", err, err)

	err = ctx.NewErrorf(c, "wrap: %w", err)
	var x ctx.Error
	test.Assert(t, errors.As(err, &x))
	t.Logf("err: %T %v", err, err)

	stack := x.Stack[0]
	t.Logf("stack: %+v", stack)

	err = ctx.WrapError(c, err)
	t.Logf("err: %T %v", err, err)

	err = fmt.Errorf("wrap more: %w", err)
	t.Logf("err: %T %v", err, err)

	err = ctx.NewErrorf(c, "new: %w", err)
	t.Logf("err: %T %v", err, err)
	test.Error(t, err)

	var cerr ctx.Error
	ok := errors.As(err, &cerr)
	test.Assert(t, ok)
	test.Error(t, cerr)
	t.Logf("err: %s", err.Error())

	test.EqualsStr(t, stack, cerr.Stack[0])
}

func TestRecoverCapturesPanicStack(t *testing.T) {
	err := func() (err error) {
		defer ctx.Recover(ctx.TODO(), &err)
		panic("boom")
	}()

	var rich ctx.Error
	test.Assert(t, errors.As(err, &rich))
	test.EqualsStr(t, "panic: boom", err.Error())
	test.Assert(t, len(rich.Stack) > 0)
	test.Assert(t, strings.Contains(rich.Stack[0], "error_test.go:"))
	test.Assert(t, !strings.Contains(rich.Stack[0], "Recover"))
}

func TestRecoverDoesNothingWithoutPanic(t *testing.T) {
	var err error
	ctx.Recover(ctx.TODO(), &err)
	test.Assert(t, err == nil)
}
