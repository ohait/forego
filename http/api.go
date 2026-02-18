package http

import (
	"io"

	"github.com/ohait/forego/api"
	"github.com/ohait/forego/api/openapi"
	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/enc"
)

type Doable interface {
	api.Op
}

type Streamable interface {
	api.StreamingOp
}

func (s *Server) MustRegisterStreamingAPI(c ctx.C, path string, obj Streamable) *openapi.PathItem {
	pi, err := s.RegisterStreamingAPI(c, path, obj)
	if err != nil {
		panic(err)
	}
	return pi
}

func (s *Server) MustRegisterAPI(c ctx.C, path string, obj Doable) *openapi.PathItem {
	pi, err := s.RegisterAPI(c, path, obj)
	if err != nil {
		panic(err)
	}
	return pi
}

func (s *Server) RegisterStreamingAPI(c ctx.C, path string, obj Streamable) (*openapi.PathItem, error) {
	handler, err := api.NewServer(c, obj)
	if err != nil {
		return nil, err
	}
	f := func(c ctx.C, in io.Reader, out func(ctx.C, any) error) error {
		req := &api.JSON{}
		if in != nil {
			err := req.ReadFrom(c, in)
			if err != nil {
				return ctx.NewErrorf(c, "can't read request body: %v", err)
			}
		} else {
			log.Infof(c, "can/t get body: %v", err)
		}
		// TODO auth

		obj, err := handler.Recv(c, req)
		if err != nil {
			return NewErrorf(c, 400, "%v", err) // receive errors are always 4xx (TODO how to handle 403?)
		}
		return obj.Stream(c, out)
	}

	if path == "" {
		return nil, ctx.NewErrorf(c, "no path to register for %T", obj)
	}

	log.Debugf(c, "registering to %q", path)
	s.handleStream(path, f)
	return handler.UpdateOpenAPI(c, s.OpenAPI, path)
}

func (s *Server) RegisterAPI(c ctx.C, path string, obj Doable) (*openapi.PathItem, error) {
	handler, err := api.NewServer(c, obj)
	if err != nil {
		return nil, err
	}
	f := func(r *Request) (any, error) {
		c := r.Context()
		req := &api.JSON{}
		if r.Body != nil {
			err := req.ReadFrom(c, r.Body)
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
		// log.Debugf(c, "API %+v", obj)

		res := &api.JSON{}
		err = handler.Send(c, obj, res)
		if err != nil {
			return nil, err
		}
		out := enc.JSON{}.Encode(c, res.Data)
		log.Debugf(c, "API response %s", out)
		return out, nil
	}

	if path == "" {
		return nil, ctx.NewErrorf(c, "no path to register for %T", obj)
	}

	log.Debugf(c, "registering to %q", path)
	s.handleRequest(path, f)
	return handler.UpdateOpenAPI(c, s.OpenAPI, path)
}
