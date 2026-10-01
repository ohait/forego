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

// Recover converts a panic in the current goroutine into a ctx.Error and stores
// the panic stack. Call it with defer and assign its result to the function's
// named error return:
//
//	func run(c ctx.C) (err error) {
//		defer ctx.Recover(c, &err)
//		// ...
//		return nil
//	}
//
// Recover is a no-op when there is no active panic. It must run in the same
// goroutine as the panic; a caller cannot recover a panic from another
// goroutine.
func Recover(c C, err *error) {
	rec := recover()
	if rec == nil || err == nil {
		return
	}
	if *err != nil {
		*err = fmt.Errorf("%w: panic: %v", *err, rec)
		return
	}
	*err = Error{
		Err:   fmt.Errorf("panic: %v", rec),
		Stack: debugStack(),
		C:     c,
	}
}

func debugStack() []string {
	frames := runtime.CallersFrames(panicCallers())
	stack := make([]string, 0, 20)
	capturingPanic := false
	for {
		frame, more := frames.Next()
		if frame.Function == "runtime.gopanic" || frame.Function == "runtime.sigpanic" {
			capturingPanic = true
			if !more {
				return stack
			}
			continue
		}
		if !capturingPanic || frame.Function == "github.com/ohait/forego/ctx.Recover" {
			if !more {
				return stack
			}
			continue
		}
		if frame.Function != "" && frame.Function != "github.com/ohait/forego/ctx.debugStack" {
			stack = append(stack, fmt.Sprintf("%s:%d", frame.File, frame.Line))
		}
		if !more {
			return stack
		}
	}
}

func panicCallers() []uintptr {
	var pcs [64]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[:n]
}

// Error is the rich error type used by Forego. It records the wrapped error,
// the stack leading to its creation, and the context active at that time so it
// can later be inspected or logged with tags intact.
type Error struct {
	Err   error    `json:"err"`
	Stack []string `json:"stack"`
	C     C        `json:"ctx"`
}

func (err Error) String() string {
	return err.Err.Error()
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
