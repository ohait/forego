package api

import "github.com/ohait/forego/ctx"

type Op interface {
	Do(c ctx.C) error
}
