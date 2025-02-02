package http

import (
	"bytes"
	"net/http"
	"time"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/utils"
)

type CallHandler struct {
	Path        string
	MaxLength   int           // default is 1Mb
	ReadTimeout time.Duration // how long to wait for the request to be completed default is 30s
	Handler     func(c ctx.C, call *Call) error
}

func (this CallHandler) readTimeout() time.Duration {
	if this.ReadTimeout > 0 {
		return this.ReadTimeout
	}
	return 30 * time.Second
}

func (this CallHandler) Register(s *Server) {
	s.mux.HandleFunc(this.Path, func(w http.ResponseWriter, r *http.Request) {
		c := r.Context()
		out, err := func() ([]byte, error) {
			c, cf := ctx.WithTimeout(c, this.readTimeout())
			defer cf()
			call := Call{
				r: r,
				w: w,
			}
			var err error
			if r.Body != nil {
				call.reqBody, err = utils.ReadAll(c, r.Body, r.Body.Close)
				if err != nil {
					return nil, NewErrorf(c, 400, "can't read body: %w", err)
				}
			}
			err = this.Handler(c, &call)
			return call.res.Bytes(), err
		}()
		if err != nil {
			w.WriteHeader(ErrorCode(err, 500))
			return
		}
		if len(out) == 0 {
			w.WriteHeader(204)
			return
		}
		w.WriteHeader(200)
		_, err = w.Write(out)
		if err != nil {
			log.Warnf(c, "writing the response: %v", err)
		}
	})

}

type Call struct {
	r *http.Request

	// only for server
	reqBody []byte
	w       http.ResponseWriter
	res     bytes.Buffer
}

func (this *Call) Write(b []byte) {
	this.res.Write(b)
}

func (this *Call) SetBody(b []byte) {
	this.res = *bytes.NewBuffer(b)
}
