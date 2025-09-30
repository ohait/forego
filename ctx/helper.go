package ctx

import (
	"context"
	"time"
)

// WithCancel returns a child context that can be cancelled along with a helper
// CancelFunc alias compatible with Forego code.
func WithCancel(c C) (C, CancelFunc) {
	c, cf := context.WithCancelCause(c)
	return c, CancelFunc(cf)
}

// WithValue mirrors context.WithValue but keeps the ctx.C alias.
func WithValue(c C, key, val any) C {
	return context.WithValue(c, key, val)
}

// TODO returns a non-nil, empty context placeholder.
func TODO() C {
	return context.TODO()
}

// Background returns a background context and a cancel helper.
func Background() (C, CancelFunc) {
	c, cf := context.WithCancelCause(context.Background())
	return c, CancelFunc(cf)
}

// Cause mirrors context.Cause to retrieve the cancellation reason.
func Cause(c C) error {
	return context.Cause(c)
}

// WithTimeout mirrors context.WithTimeout while keeping the ctx.C alias.
func WithTimeout(c C, d time.Duration) (C, func()) {
	return context.WithTimeout(c, d)
}

// WithDeadline mirrors context.WithDeadline while keeping the ctx.C alias.
func WithDeadline(c C, t time.Time) (C, func()) {
	return context.WithDeadline(c, t)
}
