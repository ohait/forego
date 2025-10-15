package ctx

import (
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
)

type trackingCtx struct {
	C
	ID   string
	Last int32
}

// WithTracking returns a context with a tracking id, if suggest is empty a new id is created
func WithTracking(c C, suggest string) C {
	if suggest == "" {
		suggest = uuid.NewString()
	}
	return trackingCtx{
		C:    WithTag(c, "tracking-id", suggest),
		ID:   suggest,
		Last: 0,
	}
}

type trackingGet struct{}

// GetTracking returns the tracking id from the context, or empty string if none exists
func GetTracking(c C) string {
	v := c.Value(trackingGet{})
	if v == nil {
		return ""
	} else {
		return v.(string)
	}
}

type trackingNext struct{}

func (c trackingCtx) Value(k any) any {
	switch k.(type) {
	case trackingNext:
		step := atomic.AddInt32(&c.Last, 1)
		return fmt.Sprintf("%s.%x", c.ID, step)
	case trackingGet:
		return c.ID
	default:
		return c.C.Value(k)
	}
}

// WithNextTracking returns a context with a new tracking id based on the previous one in the context, or a new one if none exists
// if called from a context with tracking abcd.1 it will return abcd.1.1 the first time, abcd.1.2 the second time, etc
func WithNextTracking(c C) C {
	k := c.Value(trackingNext{})
	if k == nil {
		return WithTracking(c, "")
	}
	return WithTracking(c, k.(string))
}

// Span is work in progress
func Span(c C, name string) (C, CancelFunc) {
	// TODO add opentelemetry support
	c, cf := WithCancel(c)
	k := c.Value(trackingNext{})
	if k == nil {
		k = ""
	}
	return WithTracking(c, k.(string)), cf
}
