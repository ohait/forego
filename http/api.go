package http

import (
	"bytes"
	"net/http"

	"github.com/ohait/forego/api"
	"github.com/ohait/forego/api/openapi"
	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/enc"
)

type Doable interface {
	Do(ctx.C) error
}

func (s *Server) MustRegisterAPI(c ctx.C, obj Doable) *openapi.PathItem {
	pi, err := s.RegisterAPI(c, obj)
	if err != nil {
		panic(err)
	}
	return pi
}

func (s *Server) RegisterAPI(c ctx.C, obj Doable) (*openapi.PathItem, error) {
	handler, err := api.NewServer(c, obj)
	if err != nil {
		return nil, err
	}
	f := func(c ctx.C, in []byte, r *http.Request) ([]byte, error) {
		req := &api.JSON{}
		if r.Body != nil {
			err := req.ReadFrom(c, bytes.NewBuffer(in))
			if err != nil {
				return nil, ctx.NewErrorf(c, "can't read request body: %v", err)
			}
			defer r.Body.Close()
		} else {
			log.Infof(c, "can/t get body: %v", err)
		}
		// TODO auth

		obj, err := handler.Recv(c, req)
		if err != nil {
			return nil, NewErrorf(c, 400, "%v", err) // receive errors are always 4xx (TODO how to handle 403?)
		}
		err = obj.Do(c)
		if err != nil {
			return nil, err
		}
		//log.Debugf(c, "API %+v", obj)

		res := &api.JSON{}
		err = handler.Send(c, obj, res)
		if err != nil {
			return nil, err
		}
		out := enc.JSON{}.Encode(c, res.Data)
		log.Debugf(c, "API response %s", out)
		return out, nil
	}

	urls := handler.URLs()
	if len(urls) == 0 {
		return nil, ctx.NewErrorf(c, "no URL to register for %T", obj)
	}

	for _, u := range handler.URLs() {
		log.Debugf(c, "registering to %q", u.Path)
		s.handleRequest(u.Path, f)
	}
	return handler.UpdateOpenAPI(c, s.OpenAPI)
}
