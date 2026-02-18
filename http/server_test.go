package http_test

import (
	"net/http/httptest"
	"testing"

	gohttp "net/http"

	"github.com/ohait/forego/http"
	"github.com/ohait/forego/test"
)

func TestServer(t *testing.T) {
	c := test.Context(t)

	s := http.NewServer(c)
	s.Mux().HandleFunc("/test/one", func(w gohttp.ResponseWriter, r *gohttp.Request) {
		_, _ = w.Write([]byte(`"one"`))
	})

	addr, err := s.Listen(c, "127.0.0.1:0")
	test.NoError(t, err)

	res, err := http.DefaultClient.Post(c, "http://"+addr.String()+"/test/one", []byte(`[]`))
	test.NoError(t, err)
	test.ContainsJSON(t, "one", string(res))
}

func TestServerPanic(t *testing.T) {
	c := test.Context(t)

	s := http.NewServer(c)
	s.Mux().HandleFunc("/test/panic", func(w gohttp.ResponseWriter, r *gohttp.Request) {
		panic("test panic")
	})

	addr, err := s.Listen(c, "127.0.0.1:0")
	test.NoError(t, err)

	req, err := gohttp.NewRequestWithContext(c, "POST", "http://"+addr.String()+"/test/panic", nil)
	test.NoError(t, err)

	{
		req, err := gohttp.DefaultClient.Do(req)
		test.NoError(t, err)
		test.EqualsGo(t, gohttp.StatusInternalServerError, req.StatusCode)
	}
}

func TestHandleRequestNoContent(t *testing.T) {
	c := test.Context(t)

	s := http.NewServer(c)
	s.HandleRequest("/test/no-content", func(r *gohttp.Request) (any, error) {
		return nil, nil
	})

	req := httptest.NewRequest(gohttp.MethodPost, "/test/no-content", nil)
	w := &ResponseWriter{}
	s.Mux().ServeHTTP(w, req)

	test.EqualsGo(t, gohttp.StatusOK, w.Code)
	test.EqualsStr(t, "", w.Buf.String())
}
