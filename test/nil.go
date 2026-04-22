package test

import (
	"fmt"
	"reflect"
	"testing"
)

func Nil(t testing.TB, obj any) {
	t.Helper()
	isNil(obj).argument(0, 1).true(t)
}

func NotNil(t testing.TB, obj any) {
	t.Helper()
	isNil(obj).argument(0, 1).false(t)
}

func isNil(a any) res {
	switch a := a.(type) {
	case nil:
		return res{true, "nil"}
	default:
		v := reflect.ValueOf(a)
		switch v.Kind() {
		case reflect.Slice, reflect.Map, reflect.Chan, reflect.Pointer:
			s := fmt.Sprintf("%#v", a)
			if len(s) > 100 {
				s = s[:100] + "..."
			}
			return res{v.IsNil(), s}
		default:
			return res{false, stringy{a}.String()}
		}
	}
}
