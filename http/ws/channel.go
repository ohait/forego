package ws

import (
	"sync"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/enc"
)

type Channel struct {
	Conn   *Conn
	ID     string
	byPath map[string]func(c C, n enc.Node) error
	close  func(C) error
	once   sync.Once
}

// close the channel, removing it from the connection
func (this *Channel) Close(c ctx.C) error {
	log.Infof(c, "closing channel %q", this.ID)
	this.invokeClose(c)
	return this.Conn.Send(c, Frame{
		Channel: this.ID,
		Type:    "close",
	})
}

func (this *Channel) onData(c ctx.C, f Frame) error {
	fn := this.byPath[f.Path]
	if fn == nil {
		return ctx.NewErrorf(c, "no %q for channel %q", f.Path, f.Channel)
	}
	log.Debugf(c, "ch[%q].%q(%v)", f.Channel, f.Path, f.Data)
	err := fn(C{C: c, ch: this}, f.Data)
	if err != nil {
		// log.Warnf(c, "ws: sending %v", err)
		_ = this.Conn.Send(c, Frame{
			Channel: this.ID,
			Type:    "error",
			Data:    enc.MustMarshal(c, err.Error()),
		})
	}
	return nil
}

type Frame struct {
	// dialog identifier
	Channel string `json:"channel,omitempty"`

	// routes to an object by path
	Path string `json:"path,omitempty"`

	Type string `json:"type"` // data, error, close

	Data enc.Node `json:"data,omitempty"`
}

type C struct {
	ctx.C
	ch *Channel
}

func (c C) Reply(path string, obj any) error {
	return c.ch.Conn.Send(c, Frame{
		Channel: c.ch.ID,
		Path:    path,
		Data:    enc.MustMarshal(c, obj),
	})
}

func (c C) IsClosed() bool {
	return c.ch == nil || c.ch.Conn == nil
}

// TODO(oha) should we keep it?
//func (c C) Error(obj any) error {
//	return c.ch.Conn.Send(c, Frame{
//		Channel: c.ch.ID,
//		Type:    "error",
//		Data:    enc.MustMarshal(c, obj),
//	})
//}

// Close the websocket
func (c C) Close() error {
	return c.ch.Conn.Close(c, EXIT)
}

func (this *Channel) invokeClose(c ctx.C) {
	this.once.Do(func() {
		this.Conn.byChan.Delete(this.ID)
		if this.close == nil {
			return
		}
		if err := this.close(C{
			C:  c,
			ch: this,
		}); err != nil {
			log.Warnf(c, "close channel %q: %v", this.ID, err)
		}
	})
}
