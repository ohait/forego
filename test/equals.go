package test

import (
	"fmt"
	"testing"

	"github.com/ohait/forego/utils/ast"
)

func EqualsType(t *testing.T, expect, got any) {
	t.Helper()
	equal(
		fmt.Sprintf("%T", expect),
		fmt.Sprintf("%T", got),
	).prefix("EqualsType(%s, %s)", NoQuote(ast.Assignment(0, 1)), NoQuote(ast.Assignment(0, 2))).true(t)
}

func EqualsStr(t *testing.T, expect, got string) {
	t.Helper()
	equal(expect, got).prefix("EqualsStr(%s, %s)", NoQuote(ast.Assignment(0, 1)), NoQuote(ast.Assignment(0, 2))).true(t)
}

func NotEqualsStr(t *testing.T, expect, got string) {
	t.Helper()
	equal(expect, got).prefix("NotEqualsStr(%s, %s)", NoQuote(ast.Assignment(0, 1)), NoQuote(ast.Assignment(0, 2))).false(t)
}

func equal(e, g string) res {
	if e == g {
		return res{true, e}
	} else {
		return res{false, fmt.Sprintf("%s != %s", NoQuote(e), NoQuote(g))}
	}
}

// compare using "%#v"
func EqualsGo(t *testing.T, expect, got any) {
	t.Helper()
	equalGo(expect, got).prefix("EqualsGo(%s, %s)", NoQuote(ast.Assignment(0, 1)), NoQuote(ast.Assignment(0, 2))).true(t)
}

// compare using "%#v"
func NotEqualsGo(t *testing.T, expect, got any) {
	t.Helper()
	equalGo(expect, got).prefix("NotEqualsGo(%s, %s)", NoQuote(ast.Assignment(0, 1)), NoQuote(ast.Assignment(0, 2))).false(t)
}

func equalGo(expect, got any) res {
	e := fmt.Sprintf("%#v", expect)
	g := fmt.Sprintf("%#v", got)
	if e == g {
		return res{true, e}
	} else {
		return res{false, fmt.Sprintf("%s != %s", NoQuote(e), NoQuote(g))}
	}
}

// compare using JSON
func EqualsJSON(t *testing.T, expect, got any) {
	t.Helper()
	equalJSON(expect, got).prefix("EqualsJSON(%s, %s)", ast.Assignment(0, 1), ast.Assignment(0, 2)).true(t)
}

// compare using JSON
func NotEqualsJSON(t *testing.T, expect, got any) {
	t.Helper()
	equalJSON(expect, got).prefix("NotEqualsJSON(%s, %s)", ast.Assignment(0, 1), ast.Assignment(0, 2)).false(t)
}

func equalJSON(expect, got any) res {
	e := jsonish(expect)
	g := jsonish(got)
	if e == g {
		return res{true, e}
	} else {
		return res{false, fmt.Sprintf("%s != %s", e, g)}
	}
}
