package http_test

import (
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
