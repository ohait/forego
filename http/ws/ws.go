package ws

import (
	"fmt"
	"io"
	"sync"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/enc"
	"golang.org/x/net/websocket"
)

const ( // https://www.rfc-editor.org/rfc/rfc6455#section-7.4.1
	EXIT          int = 1000
	GOING_AWAY    int = 1001
	PROTOCOL_ERR  int = 1002
	RECV_DATA_ERR int = 1003
	TOO_BIG       int = 1009
	UNEXP_COND    int = 1011
)

// low level websocket implementation
type impl interface {
	enc.ReadWriter
	Close(c ctx.C, reason int) error
}

///////////////////////////////////////////////
// implementation using /x/net/websocket

type wsImpl struct {
	m    sync.Mutex
	conn *websocket.Conn
}

var _ impl = &wsImpl{}

func (this *wsImpl) Write(c ctx.C, n enc.Node) error {
	j := enc.JSON{}.Encode(c, n)
	this.m.Lock()
	defer this.m.Unlock()
	log.Debugf(c, "ws.write: %s", j)
	return this.write(c, j)
}

// note this is not safe for multiple go routines
func (this *wsImpl) write(c ctx.C, data []byte) error {
	w, err := this.conn.NewFrameWriter(websocket.TextFrame)
	if err != nil {
		return ctx.WrapError(c, err)
	}
	ct, err := w.Write(data)
	if err != nil {
		return ctx.WrapError(c, err)
	}
	if ct != len(data) {
		return ctx.NewErrorf(c, "can't write all data: %d < %d", ct, len(data))
	}
	//log.Debugf(c, "ws sent: %s", data)
	return nil
}

func (this *wsImpl) Read(c ctx.C) (enc.Node, error) {
	j, err := this.read(c)
	if err != nil {
		return nil, err
	}
	return enc.JSON{}.Decode(c, j)
}

func (this *wsImpl) read(c ctx.C) ([]byte, error) {
	r, err := this.conn.NewFrameReader()
	if err != nil {
		return nil, err
	}
	hr := r.HeaderReader()
	if hr != nil {
		//d, _ := io.ReadAll(hr)
		//log.Debugf(c, "ws head: %x (%q)", d, d)
		_, _ = io.Copy(io.Discard, hr) // we don't care
	}
	if c.Err() != nil {
		return nil, ctx.WrapError(c, c.Err())
	}

	switch r.PayloadType() {
	case websocket.ContinuationFrame:
		return nil, ctx.NewErrorf(c, "unsupported continuation")
	case websocket.TextFrame, websocket.BinaryFrame:
	case websocket.CloseFrame:
		// TODO read message and debug?
		return nil, ctx.WrapError(c, io.EOF)
	case websocket.PingFrame, websocket.PongFrame:
		panic("unsupported ping-pong")
		/*
			            // TODO implements this
						b := make([]byte, maxControlFramePayloadLength)
						n, err := io.ReadFull(frame, b)
						if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
							return nil, err
						}
						io.Copy(ioutil.Discard, frame)
						if frame.PayloadType() == PingFrame {
							if _, err := handler.WritePong(b[:n]); err != nil {
								return nil, err
							}
						}
						return nil, nil
		*/
	default:
		panic(fmt.Sprintf("unexpected type %b", r.PayloadType()))
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, ctx.WrapError(c, err)
	}
	log.Debugf(c, "ws recv: %s", data)
	return data, nil
}

func (this *wsImpl) Close(c ctx.C, status int) error {
	err := this.conn.WriteClose(status)
	if err != nil {
		return err
	}
	return this.conn.Close()
}
