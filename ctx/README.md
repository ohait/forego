# `ctx`

We expand from `context.Context` with a few features and quality-of-life helpers:
* `ctx.C` is a thin alias over `context.Context` so call sites read `c ctx.C`
* helpers such as `WithCancel`, `WithValue`, `WithTimeout`, … keep everything in terms of `ctx.C`
* contexts are not just for cancel—they carry tags, loggers, span data, configuration and per-request overrides

## Why `c ctx.C` instead of `ctx context.Context`?

Just cosmetic, I find the traditional way distracting, especially when used everywhere.


## Tags and `ctx/log`

Each context has a bag of tags, which can added along the way. Those will be printed in each log messages, which make it particularly useful for
things like `CorrelationID`, `auth` or any context which will help debugging from a log message.

It also make it coherent when using other libraries, since they will still carry over the context.

All logging is `JSONL`, e.g.:

```
{"level":"debug","src":"github.com/ohait/forego/http/server.go:83","time":"2023-06-01T07:18:31.007411033+02:00","message":"listening to :8080","tags":{"service":"viewer"}}
```

May be wise to use a log viewer like `https://github.com/ohait/jl`   


## Rich errors: `ctx.Error`

```go
  return ctx.NewErrorf(c, "my error wrapping %w", err)
```

Having a wrapping error that provide a stack trace has proven formidable when debugging or operating.

When the logger finds a `ctx.Error` as an argument (or anything wrapping it) it will print the stack trace as part of the error message.

Use `ctx.NewErrorf` or `ctx.WrapError` to build those errors.


## Caveats

Generating stack traces is expensive in Go, so don't use wrapping errors if you expect to ignore them often. `ctx.Error` captures the
current stack and tags, which allocates and burns CPU, so avoid using it as a sentinel or control-flow marker.

## Logging

The companion package [`ctx/log`](../ctx/log/) provides JSON logging with automatic inclusion of tags, stack traces and custom payloads. Attach your own logger with `log.WithLogger(c, fn)` or use the default stderr JSONL output.

## TODO

finish setting up for opentelemetry (span) and tracking
