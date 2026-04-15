package ws

import (
	"fmt"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/enc"
	"github.com/ohait/forego/utils/sync"
)

type Channel struct {
	Conn      *Conn
	ID        string
	byPath    map[string]func(c C, n enc.Node) error
	close     func(C) error
	once      sync.Once
	reqCancel sync.Map[string, ctx.CancelFunc]
}

func (this *Channel) cancelRequest(c ctx.C, rid string) error {
	log.Warnf(c, "ws: cancelling request %q on channel %q", rid, this.ID)
	cf := this.reqCancel.Get(rid)
	if cf != nil {
		cf(fmt.Errorf("user cancelled"))
		this.reqCancel.Delete(rid)
	}
	return nil
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
		log.Warnf(c, "ws: unknown path %q for channel %q", f.Path, f.Channel)
		return this.Conn.Send(c, Frame{
			Channel: f.Channel,
			Type:    "error",
			Path:    f.Path,
			RID:     f.RID,
			Data:    enc.String("unknown path"),
		})
	}
	if this.Conn != nil && this.Conn.h != nil && this.Conn.h.Trace {
		log.Debugf(c, "ch[%q].%q(%v)", f.Channel, f.Path, f.Data)
	}
	c2, cf := ctx.WithCancel(c)
	this.reqCancel.Store(f.RID, cf)
	go func() {
		defer this.reqCancel.Delete(f.RID)
		err := fn(C{C: c2, ch: this, rid: f.RID}, f.Data)
		if err != nil {
			log.Warnf(c, "ws: %s error: %v", f.Path, err)
			if c2.Err() == nil {
				_ = this.Conn.Send(c, Frame{
					Channel: this.ID,
					Type:    "error",
					Data:    enc.MustMarshal(c, err.Error()),
					RID:     f.RID,
				})
			}
		}
	}()
	return nil
}

type Frame struct {
	// dialog identifier
	Channel string `json:"channel,omitempty"`

	// routes to an object by path
	Path string `json:"path,omitempty"`

	Type string `json:"type"` // data, error, close

	RID string `json:"rid,omitempty"` // request id for matching request and response

	Data enc.Node `json:"data,omitempty"`
}

type C struct {
	ctx.C
	ch  *Channel
	rid string
}

func (c C) Reply(path string, obj any) error {
	if c.Err() != nil {
		return c.Err() // if cancelled, don't send reply
	}
	return c.ch.Conn.SendOrClose(c, Frame{
		Channel: c.ch.ID,
		Path:    path,
		Data:    enc.MustMarshal(c, obj),
		RID:     c.rid,
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
