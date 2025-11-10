package http

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ohait/forego/api/openapi"
	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/enc"
	"github.com/ohait/forego/shutdown"
)

type Server struct {
	mux *http.ServeMux
	h   http.Handler

	// called when a request is done, by default it logs and generate metrics
	// OnResponse func(Stat)

	ready int32

	OpenAPI *openapi.Service

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If zero, the value of
	// ReadTimeout is used. If negative, or if zero and ReadTimeout
	// is zero or negative, there is no timeout.
	ReadHeaderTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	// A zero or negative value means there will be no timeout.
	WriteTimeout time.Duration
}

func (this *Server) SetReady(code int) {
	log.Infof(nil, "ready set to %d", code)
	atomic.StoreInt32(&this.ready, int32(code))
}

func (this *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			stack := string(debug.Stack())
			log.Errorf(r.Context(), "PANIC: %s\n%s", rec, stack)
			w.WriteHeader(500)
		}
	}()
	this.h.ServeHTTP(w, r)
}

func NewServer(c ctx.C) *Server {
	this := &Server{
		mux:     http.NewServeMux(),
		OpenAPI: openapi.NewService("unnamed"),
	}
	this.h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		c := r.Context()
		c = ctx.WithTag(c, "ua", r.UserAgent())
		c = ctx.WithTag(c, "path", r.URL.Path)

		w2 := &response{w, 0}
		switch w := w.(type) {
		case http.Hijacker:
			// if there is an hijacker, we need to be a bit clever
			this.mux.ServeHTTP(responseHijacker{w2, w}, r.WithContext(c))
		default:
			this.mux.ServeHTTP(w2, r.WithContext(c))
		}

		metric{
			Method: r.Method,
			Code:   w2.code,
			Path:   r.URL.Path, // TODO(oha) if we have templates, this won't work
		}.observe(time.Since(t0))
	})

	this.mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})

	this.ready = 204
	this.mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(int(atomic.LoadInt32(&this.ready)))
	})

	this.mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		c := r.Context()
		n, err := enc.Marshal(c, this.OpenAPI)
		if err != nil {
			log.Errorf(c, "can't marshal openapi: %v", err)
			w.WriteHeader(500)
			return
		}
		_, err = w.Write(enc.JSON{Indent: true}.Encode(c, n))
		if err != nil {
			log.Warnf(c, "can't send openapi: %v", err)
		}
	})
	return this
}

// deprecated use Handle and HandleFunc directly on Server
func (this Server) Mux() *http.ServeMux {
	return this.mux
}

func (this *Server) Listen(c ctx.C, addr string) (*net.TCPAddr, error) {
	s := http.Server{
		BaseContext: func(l net.Listener) context.Context {
			return ctx.WithTag(c, "http.addr", l.Addr().String())
		},
		ConnContext: func(c context.Context, conn net.Conn) context.Context {
			return ctx.WithTag(c, "http.remote", conn.RemoteAddr().String())
		},
		ConnState: func(conn net.Conn, state http.ConnState) {
			// TODO(oha): do we need to setup a limiter? if so, this is to know when any hijacker kicks in
		},
		ReadTimeout:       this.ReadTimeout,
		ReadHeaderTimeout: this.ReadHeaderTimeout,
		WriteTimeout:      this.WriteTimeout,

		Handler: this,
	}
	if addr == "" {
		addr = ":http"
	}
	log.Debugf(c, "listening to %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	ch := make(chan error, 1)

	// we start the server in a goroutine
	go func() {
		defer shutdown.Hold().Release()
		// we wrap the listener, so the first call to Accept() will write nil to the error channel
		l := &listener{
			Listener: ln,
			f: func() {
				ch <- nil
			},
		}
		err := s.Serve(l)
		if err != nil {
			log.Warnf(c, "listen(%q) %v", addr, err)
		}
	}()

	go func() {
		<-shutdown.Started()
		c, cf := ctx.WithTimeout(c, 30*time.Second)
		defer cf()
		_ = s.Shutdown(c)
	}()

	// blocks until either an error, or the first Accept() call happen
	return ln.Addr().(*net.TCPAddr), <-ch
}

// wrapper for a net listener
type listener struct {
	net.Listener
	once sync.Once
	f    func()
}

func (this *listener) Accept() (net.Conn, error) {
	this.once.Do(this.f)
	return this.Listener.Accept()
}

// Setup the given streaming function, ignore the method, request body can be nil
func (this *Server) HandleStream(path string, f StreamFunc) *openapi.PathItem {
	this.handleStream(path, f)
	return this.makePathItem(path)
}

// Setup the given request as JSON, and add it to `s.OpenAPI` for the given path as POST, returns the openapi.PathItem
func (this *Server) HandleRequest(path string, f func(c ctx.C, in []byte, r *http.Request) ([]byte, error)) *openapi.PathItem {
	this.handleRequest(path, f)
	return this.makePathItem(path)
}

func (this *Server) makePathItem(path string) *openapi.PathItem {
	pi := &openapi.PathItem{
		RequestBody: &openapi.RequestBody{
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Type: "object", // we assume it's an object
					},
				},
			},
		},
		Responses: map[string]openapi.Response{
			"200": {
				Content: map[string]openapi.Content{
					"application/json": {
						Schema: &openapi.Schema{
							Type: "object",
						},
					},
				},
			},
		},
	}
	this.OpenAPI.Paths[path] = &openapi.Path{
		Post: pi,
	}
	return pi
}

// a streaming function with an option request body `in` and a function which sends chunks to the client
type StreamFunc func(c ctx.C, in io.Reader, emit func(c ctx.C, obj any) error) error

func (this *Server) handleStream(path string, f StreamFunc) {
	this.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := r.Context()
		flusher, ok := w.(http.Flusher)
		if !ok {
			log.Errorf(c, "can't cast to http.Flusher: %T", w)
			w.WriteHeader(500)
			return
		}
		sent := 0
		if r.Body != nil {
			defer r.Body.Close()
		}
		err := f(c, r.Body, func(c ctx.C, obj any) error {
			if c.Err() != nil {
				return c.Err()
			}
			json, err := enc.MarshalJSON(c, obj)
			if err != nil {
				return err
			}
			sent += len(json) + 1
			w.Write(json)
			w.Write([]byte("\n"))
			flusher.Flush()
			return nil
		})
		if err == nil {
			if sent == 0 {
				w.WriteHeader(204)
			}
			return
		}
		// error handling
		if sent == 0 {
			log.Warnf(c, "error before streaming: %v", err)
			switch err := err.(type) {
			case Error:
				w.WriteHeader(err.Code)
				w.Write([]byte(err.Err.Error()))
			default:
				w.WriteHeader(500) // TODO is there a way to send a different code?
			}
			return
		}
		log.Errorf(c, "error while streaming: %v", err)
	})
}

func (this *Server) handleRequest(path string, f func(c ctx.C, in []byte, r *http.Request) ([]byte, error)) {
	this.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := r.Context()
		out, err := func() ([]byte, error) {
			var in []byte
			var err error
			if r.Body != nil {
				in, err = io.ReadAll(r.Body)
				if err != nil {
					return nil, NewErrorf(c, 400, "can't read body: %w", err)
				}
			}
			return f(c, in, r)
		}()
		if err != nil {
			tid := ctx.GetTracking(c)
			log.Warnf(c, "http: %v", err)
			code := ErrorCode(err, 500)
			w.WriteHeader(code)
			if code < 500 {
				j, _ := json.Marshal(map[string]any{
					"error":    err.Error(),
					"tracking": tid,
				})
				_, _ = w.Write(j)
			} else {
				j, _ := json.Marshal(map[string]any{
					// NO CLIENT REPORTING FOR INTERNAL ERRORS (due to security reason, we don't want to leak information about internals)
					"tracking": tid,
				})
				_, _ = w.Write(j)
			}
			return
		}

		if len(out) == 0 {
			// no response content
			w.WriteHeader(204)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if len(out) > 16*1024 && strings.Contains(r.Header.Get("Accept"), "gzip") { // TODO ugly parsing, but good enough for now
			w.Header().Add("Content-Encoding", "gzip")
			w2 := gzip.NewWriter(w)
			_, err = w2.Write(out)
			w2.Close()
			log.Debugf(c, "sending gzip %d", len(out))
		} else {
			_, err = w.Write(out)
		}
		if err != nil {
			log.Warnf(c, "writing the response: %v", err)
		}
	})
}
