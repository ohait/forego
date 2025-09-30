package ctx

import (
	"errors"
	"fmt"
	"runtime"
)

// NewErrorf formats an error and ensures it carries a stack trace plus the
// originating context. It behaves like fmt.Errorf but automatically wraps the
// result in ctx.Error so log messages can include tracebacks and tags.
func NewErrorf(c C, f string, args ...any) error {
	return maybeWrap(c, fmt.Errorf(f, args...))
}

// WrapError attaches a stack trace and context to err unless it already holds
// that information. It is safe to call with nil.
func WrapError(c C, err error) error {
	if err == nil {
		return nil
	}
	return maybeWrap(c, err)
}

// Error is the rich error type used by Forego. It records the wrapped error,
// the stack leading to its creation, and the context active at that time so it
// can later be inspected or logged with tags intact.
type Error struct {
	Err   error    `json:"err"`
	Stack []string `json:"stack"`
	C     C        `json:"ctx"`
}

// Error implements the error interface by forwarding to the wrapped error.
func (err Error) Error() string {
	return err.Err.Error()
}

// Unwrap returns the underlying error so errors.Is / errors.As keep working.
func (err Error) Unwrap() error {
	return err.Err
}

// Is reports whether err matches Error or the wrapped value, allowing callers
// to detect Forego errors via errors.Is.
func (this Error) Is(err error) bool {
	switch err.(type) {
	case *Error, Error:
		return true
	default:
		return errors.Is(this.Err, err)
	}
}

func maybeWrap(c C, err error) error {
	if errors.Is(err, Error{}) {
		return err // already wrapped
	}
	if errors.Is(err, &Error{}) {
		return err // already wrapped
	}
	return Error{
		Err:   err,
		Stack: stack(2, 100),
		C:     c,
	}
}

func stack(above, max int) []string {
	stack := make([]string, 0, 20)
	for len(stack) < max {
		_, file, line, ok := runtime.Caller(above + 1)
		if !ok {
			return stack
		}
		stack = append(stack, fmt.Sprintf("%s:%d", file, line))
		above++
	}
	return stack
}

// JSON behaves like json.RawMessage while remaining printable in log tags.
type JSON []byte

// MarshalJSON returns the raw bytes, keeping JSON compatibility.
func (this JSON) MarshalJSON() ([]byte, error) {
	return this, nil
}

// UnmarshalJSON stores the raw JSON payload.
func (this *JSON) UnmarshalJSON(j []byte) error {
	*this = j
	return nil
}

// String renders the payload verbatim, handy for logs.
func (this JSON) String() string {
	return string(this)
}
