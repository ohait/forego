# forego

Build Go JSON/WebSocket services with tagged logging, auto-wired APIs, and test-first ergonomics.

## Why Forego
- Tagged JSON logs keep every request’s story without coupling handlers to a logger. Attach tags with `ctx.WithTag`, and `ctx/log` takes care of consistently emitting them.
- The `http`, `api`, and `test` packages form a feedback loop: define a Go struct once, expose it over HTTP/WebSocket, and exercise it in your tests without binding boilerplate.
- Batteries stay optional. Supporting packages like `enc`, `shutdown`, and `utils/prom` plug in when you need them without locking you into a monolith.

## Quickstart
1. Install the module:
   ```bash
   go get github.com/ohait/forego@latest
   ```
2. Wire an API, HTTP server, and graceful shutdown:
   ```go
   package main

   import (
       "regexp"

       fctx "github.com/ohait/forego/ctx"
       flog "github.com/ohait/forego/ctx/log"
       fhttp "github.com/ohait/forego/http"
       "github.com/ohait/forego/shutdown"
   )

   type WordFilter struct {
       Blacklist *regexp.Regexp

       In  string `api:"in,required"`
       Out string `api:"out"`
   }

   func (wf *WordFilter) Do(c fctx.C) error {
       wf.Out = wf.Blacklist.ReplaceAllString(wf.In, "***")
       return nil
   }

   func main() {
       c, cf := fctx.Background()
       defer cf(nil)

       c = fctx.WithTag(c, "service", "wordfilter")

       srv := fhttp.NewServer(c)
       srv.MustRegisterAPI(c, "/api/wordfilter/v1", &WordFilter{
           Blacklist: regexp.MustCompile(`(foo|bar)`),
       })

       addr, err := srv.Listen(c, "127.0.0.1:8080")
       if err != nil {
           panic(err)
       }
       flog.Infof(c, "listening on %s", addr.String())

       shutdown.WaitForSignal(c, cf)
   }
   ```
3. Call the endpoint; the OpenAPI definition is already available at `/openapi.json`:
   ```bash
   curl -s -X POST localhost:8080/api/wordfilter/v1 \
     -H 'content-type: application/json' \
     -d '{"in":"foo and friends"}'
   ```
4. Test the same struct without HTTP glue:
   ```go
   import (
       "regexp"
       "testing"

       "github.com/ohait/forego/api"
       "github.com/ohait/forego/test"
    )

    func TestWordFilter(t *testing.T) {
        c := test.Context(t)
        out := api.Test(c, &WordFilter{
            Blacklist: regexp.MustCompile(`(foo|bar)`),
            In:        "foo and bar",
        })

        test.EqualsStr(t, "*** and ***", out.Out)
    }
    ```

## Architecture Overview
- `ctx` / `ctx/log`: wrap `context.Context` so tags, rich errors, and JSON logs travel together; handlers don’t have to know who consumes the logs.
- `http`, `api`, `http/ws`: map structs to REST and WebSocket endpoints, emit OpenAPI, and reuse the same types in business logic and tests.
- `test`: AST-aware assertions that reuse your API structs directly, producing readable success and failure output.
- `enc`: an intermediate JSON representation that keeps parsing efficient when you need custom coercion.
- `shutdown`, `utils/prom`, `storage`: supporting utilities for graceful shutdown, lightweight metrics, and simple storage patterns.

## Observability by Default
```go
c = fctx.WithTag(c, "tracking", "a1b2c3")
flog.Infof(c, "completed request")
```
Produces a JSONL entry similar to:
```
{"level":"info","message":"completed request","tags":{"service":"wordfilter","tracking":"a1b2c3"}}
```
Every handler in `forego/http` automatically enriches tags with user agent, path, remote address, and exposes the same metadata to your own log calls.

## Package Tour
- [api](./api/) — bind Go structs to RPC-style operations, generate OpenAPI, and provide helpers to test them directly.
- [enc](./enc/) — flexible JSON intermediate forms (`enc.Node`, `enc.Map`, `enc.Pairs`) that avoid `json.RawMessage` gymnastics.
- [http](./http/) — production-grade HTTP server with automatic request tagging, gzip, OpenAPI serving, and API registration.
- [http/ws](./http/ws/) — WebSocket RPC bindings that reuse the same struct/tag approach as REST.
- [ctx](./ctx/) — context helpers, tagged metadata, structured logging, and rich error wrappers.
- [shutdown](./shutdown/) — graceful signal handling with hold/release coordination.
- [test](./test/) — expressive assertions powered by source analysis.
- [utils/prom](./utils/prom/) — lightweight Prometheus-compatible metrics.
- [utils](./utils/) — shared helpers (AST, caches, lists, sync) used across packages.

## When to Use / When to Wait
- Use it when you’re building Go JSON or WebSocket services and want consistent logging, automatic documentation, and tests that stay close to your business logic.
- Wait if you need a polished, stable framework or non-Go integrations—this project is still evolving and APIs may move.

Further details live in the package READMEs linked above.
