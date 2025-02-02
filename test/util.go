package test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ohait/forego/enc"
	"github.com/ohait/forego/utils/ast"
)

// helper, returns the jsonish value as string, or an error as string
// just to make tests easier to manage
// NOTE: we assume the error message is never a valid jsonish, so there is no ambiguity
func jsonish(v any) string {
	switch x := v.(type) {
	case json.RawMessage:
		_ = json.Unmarshal(x, &v) // make it into any, so we can marshal it again and canonicalize it
	case []byte:
		if json.Valid(x) {
			_ = json.Unmarshal(x, &v) // make it into any, so we can marshal it again and canonicalize it
		}
	case string:
		if json.Valid([]byte(x)) {
			_ = json.Unmarshal([]byte(x), &v) // make it into any, so we can marshal it again and canonicalize it
		}
	}
	//j, err := json.Marshal(v)
	j, err := enc.MarshalJSON(nil, v)
	if err != nil {
		return err.Error()
	}
	return string(j)
}

type res struct {
	succeed bool
	msg     string
}

// retrieve the argument of the function
func (res res) argument(above, argNum int) res {
	call, _, _ := ast.Caller(above + 1)
	res.msg = call.Args[argNum].Src + ": " + res.msg
	return res
}

// retrieve the argument of the function, and follow where it was assigned if it's a simple variable
func (res res) assignment(above, argNum int) res { // nolint:unused
	res.msg = ast.Assignment(above+1, argNum) + ": " + res.msg
	return res
}

func (res res) prefix(f string, args ...any) res {
	res.msg = fmt.Sprintf(f, args...) + " ⮕  " + res.msg
	return res
}

// expect true
func (res res) true(t *testing.T) {
	t.Helper()
	if res.succeed {
		OK(t, "%s", res.msg)
	} else {
		Fail(t, "%s", res.msg)
	}
}

// expect false
func (res res) false(t *testing.T) {
	t.Helper()
	if res.succeed {
		Fail(t, "%s", res.msg)
	} else {
		OK(t, "%s", res.msg)
	}
}

// using this will make the string representation more human
type stringy struct {
	any
}

func (this stringy) String() string {
	switch v := this.any.(type) {
	case json.RawMessage:
		return string(v)
	case []byte:
		if utf8.Valid(v) {
			return Quote(string(v))
		} else {
			return fmt.Sprintf("%#v", v)
		}
	case string:
		return Quote(v)
	case fmt.Stringer:
		return Quote(v.String())
	default:
		return fmt.Sprintf("%#v", v)
	}
}

func NoQuote(s string) string { return s }

func Quote(s string) string {
	r := strings.NewReplacer(
		"`", "\\`",
		`\`, `\\`,
		"\t", `\t`,
		"\r", `\r`,
		"\n", `\v`,
	)
	return "`" + r.Replace(s) + "`"
}
