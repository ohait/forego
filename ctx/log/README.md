# `ctx/log`

Structured logging utilities built on top of `ctx.C`.

* Log messages are emitted as JSON Lines (`Line.JSON()`) and always include the
  tags stored in the originating context.
* `log.Errorf`, `log.Warnf`, `log.Infof`, and `log.Debugf` mirror the standard
  library's formatting helpers but understand `ctx.Error` values so stack traces
  and tagged contexts are captured automatically.
* Attach a custom sink with `log.WithLogger(ctx, func(Line))`; tests can use
  `log.WithLoggerAndHelper` to hook into `t.Helper()` and keep backtraces clean.
* Implement `log.Loggable` on your types to tweak how they appear in structured
  logs and to add/remove tags dynamically.

See the root [`ctx`](../) package README for an overview of tagging and error
helpers that work hand-in-hand with the logger.
