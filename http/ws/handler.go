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

	// Called when shutdown.Started()
	OnShutdown func(c ctx.C, conn *Conn)

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

// MustRegister is like Register but panics on error.
func (this *Handler) MustRegister(c ctx.C, obj any) *Handler {
	err := this.Register(c, obj)
	if err != nil {
		panic(err)
	}
	return this
}

// Register registers a struct as a WebSocket handler, making its methods available as RPC endpoints.
//
// The obj parameter must be a struct or pointer to struct. Non-zero field values are shallow-copied
// to each handler instance created per WebSocket channel.
//
// # Special Methods
//
// Init - Constructor called when a channel opens:
//
//	func (h *Handler) Init(c ws.C) error
//	func (h *Handler) Init(c ws.C, arg T) error
//
// The optional argument is deserialized from the opening frame's data field.
//
// Close - Destructor called when a channel closes:
//
//	func (h *Handler) Close(c ws.C) error
//
// Called once for cleanup. Must not accept arguments beyond the context.
//
// # Handler Methods
//
// Regular methods become RPC endpoints with signatures:
//
//	func (h *Handler) MethodName(c ws.C) error
//	func (h *Handler) MethodName(c ws.C, arg T) error
//
// The first parameter must be ws.C. An optional second parameter is deserialized from the frame data.
// Method names are exposed as camelCase (first letter lowercased). Methods without ws.C as the first
// parameter are ignored.
//
// # Execution Flow
//
//  1. Client opens channel: {Channel: "id", Path: "handlerName", Type: "open", Data: ...}
//  2. Handler struct is instantiated with copied field values
//  3. Init method called with opening frame data (if exists)
//  4. Method calls routed via Path to handler methods
//  5. Returns sent as: {Channel: "id", Path: "methodName", Type: "return"}
//  6. On close, Close method invoked for cleanup
//
// # Example
//
//	type Counter struct {
//	    MinAmt int  // Copied to each instance
//	    Ct     int
//	}
//
//	func (c *Counter) Init(ctx ws.C, startAmt int) error {
//	    c.Ct = startAmt
//	    return nil
//	}
//
//	func (c *Counter) Inc(ctx ws.C, amt int) error {
//	    if amt < c.MinAmt {
//	        return fmt.Errorf("amount too small")
//	    }
//	    c.Ct += amt
//	    return ctx.Reply("count", c.Ct)
//	}
//
//	func (c *Counter) Close(ctx ws.C) error {
//	    // cleanup
//	    return nil
//	}
//
// See reflect_test.go and http_test.go for more examples.
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
