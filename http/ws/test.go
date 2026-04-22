package ws

import (
	"io"

	"github.com/google/uuid"
	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/enc"
)

type TestClient struct {
	conn *Conn
	ws   *testWS
}

type testWS struct {
	byChan map[string]func(ctx.C, Frame) error
	inbox  chan Frame
	closed bool
}

var _ impl = &testWS{}

func (this *testWS) Close(c ctx.C, reason int) error {
	this.closed = true
	return nil
}

func (this *testWS) Read(c ctx.C) (enc.Node, error) {
	select {
	case f, ok := <-this.inbox:
		if !ok {
			return nil, io.EOF
		}
		log.Debugf(c, "%T.read() %+v", this, f)
		return enc.Marshal(c, f)
	case <-c.Done():
		return nil, c.Err()
	}
}

func (this *testWS) Write(c ctx.C, n enc.Node) error {
	if this.closed {
		return io.ErrClosedPipe
	}
	var f Frame
	if err := enc.Unmarshal(c, n, &f); err != nil {
		return ctx.WrapError(c, err)
	}

	h := this.byChan[f.Channel]
	if h == nil {
		return ctx.NewErrorf(c, "unknown channel %q", f.Channel)
	}
	return h(c, f)
}

// create a local websocket loop and return a client connected to it
func (this *Handler) NewTest(c ctx.C) *TestClient {
	ws := &testWS{
		byChan: map[string]func(ctx.C, Frame) error{},
		inbox:  make(chan Frame, 10),
	}
	conn := &Conn{
		h:  this,
		ws: ws,
	}
	go conn.Loop(c) // nolint

	return &TestClient{
		conn: conn,
		ws:   ws,
	}
}

func (this *TestClient) Send(c ctx.C, f Frame) error {
	select {
	case this.ws.inbox <- f:
		log.Debugf(c, "%T.send() %+v", this, f)
		return nil
	case <-c.Done():
		return c.Err()
	}
}

func (this *TestClient) Close(c ctx.C) error {
	return this.conn.Close(c, 1000)
}

type TestChannel struct {
	cli      *TestClient
	id       string
	ech      map[string]chan error
	cb       map[string]func(ctx.C, Frame) error
	fallback func(ctx.C, Frame) error
}

func (this *TestChannel) Close(c ctx.C) error {
	return this.cli.Close(c)
}

func (this *TestChannel) onData(c ctx.C, f Frame) error {
	switch f.Type {
	case "return", "error":
		ech := this.ech[f.RID]
		if ech == nil {
			return this.fallback(c, f)
		}
		var err error
		if f.Data != nil {
			err = ctx.NewErrorf(c, "remote: %s", f.Data)
		}
		if ech != nil {
			ech <- err
		}
		return nil
	case "":
		cb := this.cb[f.RID]
		if cb != nil {
			return cb(c, f)
		}
		return this.fallback(c, f)
	default:
		log.Warnf(c, "unknown type %s", f.Type)
		return nil
	}
}

func (this *TestChannel) Send(c ctx.C, f Frame) error {
	return this.cli.Send(c, f)
}

func (this *TestChannel) SendData(c ctx.C, path string, data any) error {
	n, err := enc.Marshal(c, data)
	if err != nil {
		return err
	}
	return this.Send(c, Frame{Path: path, Data: n})
}

func (this *TestChannel) Request(c ctx.C, path string, args any,
	onData func(ctx.C, Frame) error,
) error {
	rid := uuid.NewString()
	ech := make(chan error, 1)
	this.ech[rid] = ech
	this.cb[rid] = onData
	err := this.cli.Send(c, Frame{
		RID:     rid,
		Channel: this.id,
		Path:    path,
		Data:    enc.MustMarshal(c, args),
	})
	if err != nil {
		return err
	}
	select {
	case err := <-ech:
		return err
	case <-c.Done():
		return c.Err()
	}
}

// open a channel, and return a test channel handler
// the handler can be used to call remote functions
func (this *TestClient) Open(c ctx.C, path string, data any,
	onData func(ctx.C, Frame) error,
) (*TestChannel, error) {
	ch := &TestChannel{
		cli:      this,
		id:       uuid.NewString(),
		ech:      map[string]chan error{},
		cb:       map[string]func(ctx.C, Frame) error{},
		fallback: onData,
	}
	this.ws.byChan[ch.id] = ch.onData
	return ch, this.Send(c, Frame{
		Type:    "open",
		Channel: ch.id,
		Path:    path,
		Data:    enc.MustMarshal(c, data),
	})
}

// experimental
// open a channel, and return a ws.C for that channel and a inbox where all the responses will be added to
func (this *TestClient) NewContext(c ctx.C, size int) (C, <-chan Frame) {
	outbox := make(chan Frame, size)
	ch := Channel{
		Conn: this.conn,
		ID:   uuid.NewString(),
	}
	this.ws.byChan[ch.ID] = func(c ctx.C, f Frame) error {
		select {
		case outbox <- f:
			return nil
		case <-c.Done():
			return c.Err()
		}
	}
	return C{
		C:  c,
		ch: &ch,
	}, outbox
}
