package test

import (
	"errors"
	"testing"

	"github.com/ohait/forego/utils/ast"
)

// var ok = "✅  "
var (
	ok = "  ✔ "
	ko = "❌  "
)

func OK(t testing.TB, f string, args ...any) {
	t.Helper()
	t.Logf(ok+f, args...)
}

func Fail(t testing.TB, f string, args ...any) {
	t.Helper()
	t.Fatalf(ko+f, args...)
}

func Assert(t testing.TB, cond bool) {
	t.Helper()
	if cond {
		OK(t, "%s", stringy{ast.Assignment(0, 1)})
	} else {
		Fail(t, "%s", stringy{ast.Assignment(0, 1)})
	}
}

var ExpectedError = errors.New("expected error")
