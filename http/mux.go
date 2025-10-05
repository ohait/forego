package http

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ohait/forego/api/openapi"
	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/enc"
	"github.com/ohait/forego/utils"
)

func HandleRequest[Req any](
	c ctx.C,
	s *Server,
	path string,
	f func(ctx.C, Req, http.ResponseWriter) error, // NOTE: the ctx.C comes from the server
) *openapi.PathItem {
	oa := s.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := r.Context()
		if r.Body == nil {
			http.Error(w, "missing request body", 400)
			return
		}
		defer r.Body.Close()
		var req Req
		if r.ContentLength > 10*1024*1024 {
			http.Error(w, "request body too large", 400)
			return
		}
		body := make([]byte, r.ContentLength)
		_, err := io.ReadFull(r.Body, body)
		if err != nil {
			http.Error(w, "can't read request body: "+err.Error(), 400)
			return
		}
		err = enc.UnmarshalJSON(c, body, &req)
		if err != nil {
			http.Error(w, "can't decode request body: "+err.Error(), 400)
			return
		}
		log.Debugf(c, "decoded request: %+v", req)

		// call the function
		err = f(c, req, w)
		if err != nil {
			log.Warnf(c, "request: %v", err)
			http.Error(w, "request: "+err.Error(), 500)
		}
	})
	var req Req
	oa.Post = &openapi.PathItem{
		Summary: "Handle " + path,
		RequestBody: &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: s.OpenAPI.MustSchemaFromType(c, req),
				},
			},
		},
	}
	return oa.Post
}

func (this *Server) HandleFunc(path string, h http.HandlerFunc) *openapi.Path {
	return this.Handle(path, http.HandlerFunc(h))
}

func (this *Server) Handle(path string, h http.Handler) *openapi.Path {
	this.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		c := r.Context()
		c = ctx.WithTag(c, "ua", r.UserAgent())
		c = ctx.WithTag(c, "path", r.URL.Path)

		w2 := &response{w, 0}
		switch w := w.(type) {
		case http.Hijacker:
			// if there is an hijacker, we need to be a bit clever
			h.ServeHTTP(responseHijacker{w2, w}, r.WithContext(c))
		default:
			h.ServeHTTP(w2, r.WithContext(c))
		}

		metric{
			Method: r.Method,
			Code:   w2.code,
			Path:   path,
		}.observe(time.Since(t0))
	})

	p := &openapi.Path{}
	this.OpenAPI.Paths[path] = p
	return p
}

type response struct {
	http.ResponseWriter
	code int
}

type responseHijacker struct {
	*response
	hijacker http.Hijacker
}

var (
	_ http.Hijacker = &responseHijacker{}
	_ http.Flusher  = responseHijacker{}
)

func (r *response) WriteHeader(code int) {
	if r.code != 0 {
		stack := utils.Stack(1, 10)
		log.Warnf(nil, "duplicate WriteHeader() at %s", strings.Join(stack, "\n"))
		return
	}
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *response) Write(b []byte) (int, error) {
	if r.code == 0 {
		r.code = 200
	}
	return r.ResponseWriter.Write(b)
}

func (r responseHijacker) Flush() {
	r.hijacker.(http.Flusher).Flush()
}

func (r responseHijacker) Hijack() (conn net.Conn, rw *bufio.ReadWriter, err error) {
	return r.hijacker.Hijack()
}
