package ws_test

import (
	"testing"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/http/ws"
	"github.com/ohait/forego/test"
)

type Example struct {
	// some pointless statistics
	byLength map[int]float64
}

func (this *Example) Init(c ws.C) error {
	this.byLength = map[int]float64{}
	return nil
}

// add multiple words to our statistics
func (this *Example) AddWords(c ws.C, req []struct {
	Word   string   `json:"word"`
	Weight *float64 `json:"weight,omitempty"`
}) error {
	for _, entry := range req {
		l := len(entry.Word)
		w := 1.0
		if entry.Weight != nil {
			w = *entry.Weight
		}
		this.byLength[l] += w
	}
	return nil
}

func (this *Example) Stats(c ws.C) error {
	return c.Reply("stats", this.byLength)
}

func (this *Example) ByLength(c ws.C, len int) error {
	return c.Reply("stats", this.byLength[len])
}

func TestExample(t *testing.T) {
	c := test.Context(t)
	h := ws.Handler{}
	test.NoError(t, h.Register(c, &Example{}))
	cli := h.NewTest(c)

	send, err := cli.Open(c, "example", nil, func(c ctx.C, f ws.Frame) error {
		switch f.Type {
		case "close":
			t.Logf("recv CLOSED")
		default:
			t.Logf("recv %+v", f.Data)
		}
		return nil
	})
	test.NoError(t, err)
	test.NoError(t, send(c, "addWords", []any{
		map[string]any{
			"word": "foo",
		},
		map[string]any{
			"word":   "bar",
			"weight": 3.14,
		},
	}))
	test.NoError(t, send(c, "stats", nil))
	test.NoError(t, cli.Close(c))
}
