package ws

import (
	"fmt"
	"net/http"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/enc"
	"github.com/ohait/forego/shutdown"
	"github.com/ohait/forego/utils/sync"
	"golang.org/x/net/websocket"
)

type Handler struct {
	// Trace enables verbose frame and dispatch logging when true.
	Trace bool

	byPath sync.Map[string, func(ctx.C, *Conn, Frame) error]
}

// return a websocket.Server which can be used as an http.Handler
// Note: it sets a default Handshake handler which accept any requests,
// you might need to change it if you need to control the `Origin` header.
func (this *Handler) Server() websocket.Server {
	x := websocket.Server{
		Handler: websocket.Handler(func(conn *websocket.Conn) {
			c := conn.Request().Context()
			c, cf := ctx.Span(c, "ws")
			defer cf(nil)

			defer shutdown.Hold().Release()

			// defer metrics.WS{Path: path}.Start().End(c)
			ws := Conn{
				h: this,
				ws: &wsImpl{
					conn:  conn,
					trace: this.Trace,
				},
			}
			defer ws.Close(c, 1000)
			err := ws.Loop(c)
			if err != nil {
				log.Debugf(c, "loop: %v", err)
			}
		}),
		Handshake: func(config *websocket.Config, req *http.Request) (err error) {
			config.Origin, err = websocket.Origin(config, req)
			if err == nil && config.Origin == nil {
				return fmt.Errorf("null origin")
			}
			return err
		},
	}
	return x
}

func (this *Handler) MustRegister(c ctx.C, obj any) *Handler {
	err := this.Register(c, obj)
	if err != nil {
		panic(err)
	}
	return this
}

func (this *Handler) Register(c ctx.C, obj any) error {
	var b builder
	err := b.inspect(c, obj)
	if err != nil {
		return err
	}
	this.byPath.Store(b.name, func(c ctx.C, conn *Conn, f Frame) error {
		log.Debugf(c, "new %q...", b.name)
		ch := &Channel{
			Conn:   conn,
			byPath: map[string]func(c C, n enc.Node) error{},
			ID:     f.Channel,
		}
		conn.byChan.Store(ch.ID, ch)
		obj := b.build(C{
			C:  c,
			ch: ch,
		}, f.Data)
		log.Debugf(c, "new %+v", obj)
		return nil
	})
	return nil
}
