// Package ctx supplements Go's context package with helpers for tagging, error
// wrapping, and logging integration used throughout Forego.
package ctx

import "context"

// C is a light alias for Go's context.Context used throughout Forego code so
// call sites can read `c ctx.C` instead of the longer `ctx context.Context`.
// It behaves exactly like the standard context and can be passed anywhere a
// context.Context is required.
type C context.Context

// CancelFunc mirrors context.CancelCauseFunc and is returned by helpers such as
// context.WithCancelCause. The Exit helper makes it obvious that we are done
// with the context by cancelling it with a nil cause.
type CancelFunc context.CancelCauseFunc

// Exit cancels the context without an error cause.
func (f CancelFunc) Exit() {
	f(nil)
}
