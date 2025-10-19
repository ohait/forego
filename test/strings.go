package test

import (
	"fmt"
	"strings"
	"testing"
)

func NotContains(t testing.TB, str any, pattern string) {
	t.Helper()
	contains(fmt.Sprintf("%v", str), pattern).false(t)
}

func Contains(t testing.TB, str any, pattern string) {
	t.Helper()
	contains(fmt.Sprintf("%v", str), pattern).true(t)
}

func ContainsGo(t testing.TB, obj any, pattern string) {
	t.Helper()
	contains(fmt.Sprintf("%#v", obj), pattern).true(t)
}

// check if the json of obj contains pattern
func ContainsJSON(t testing.TB, obj any, pattern string) {
	t.Helper()
	s := jsonish(obj)
	contains(s, pattern).true(t)
}

// check if the json of obj does NOT contains pattern
func NotContainsJSON(t testing.TB, obj any, pattern string) {
	t.Helper()
	s := jsonish(obj)
	contains(s, pattern).false(t)
}

func contains(s, pattern string) res {
	if strings.Contains(s, pattern) {
		return res{true, Quote(s)}
	} else {
		return res{false, fmt.Sprintf("%s not in %s", Quote(pattern), Quote(s))}
	}
}
