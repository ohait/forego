package api_test

import (
	"regexp"
	"testing"

	"github.com/ohait/forego/api"
	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/http"
	"github.com/ohait/forego/test"
)

type WordFilter struct {
	Blacklist *regexp.Regexp

	In    string `api:"in,required" json:"in"`
	Out   string `api:"out" json:"out"`
	Count int    `api:"out" json:"count"`
}

func (wf *WordFilter) Do(c ctx.C) error {
	wf.Out = wf.Blacklist.ReplaceAllStringFunc(wf.In, func(bad string) string {
		wf.Count++
		return "***"
	})
	return nil
}

func TestWordFilter(t *testing.T) {
	c := test.Context(t)
	re := regexp.MustCompile(`(bad|worse|worst)`)
	out := api.Test(c, &WordFilter{
		Blacklist: re,
		In:        "ok, bad or worse",
	})
	test.EqualsStr(t, "ok, *** or ***", out.Out)
	test.EqualsGo(t, 2, out.Count)
}

func exampleWordFilter(c ctx.C) { // nolint
	s := http.NewServer(c)
	_, _ = s.RegisterAPI(c, "/wc", &WordFilter{
		Blacklist: regexp.MustCompile(`(foo|bar)`), // this will be copied by ref for each request
	})
}

func exampleHandler(c ctx.C) error { // nolint
	h, err := api.NewHandler(c, &WordFilter{})
	if err != nil {
		return err
	}
	ser := h.Server()

	onRequest := func(c ctx.C, req api.ServerRequest, res api.ServerResponse) error {
		// un marshal the request into a new *WordFilter
		op, err := ser.Recv(c, req)
		if err != nil {
			return err
		}

		// call *WordFilter.Do()
		err = op.Do(c)
		if err != nil {
			return err
		}

		// marshal back the response
		return ser.Send(c, op, res)
	}

	_ = onRequest

	return nil
}
