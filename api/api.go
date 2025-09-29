package api

import "github.com/ohait/forego/ctx"

type Op interface {
	Do(c ctx.C) error
}

type StreamingOp interface {
	Stream(c ctx.C, emit func(ctx.C, any) error) error
}
