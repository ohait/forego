package utils

import (
	"bytes"
	"io"

	"github.com/ohait/forego/ctx"
)

// ReadAll() but honors the ctx.C cancel
// Note: it uses an internal go-routine, which will block on read
// If the close function is passed, it will be called when read is exhausted.
func ReadAll(c ctx.C, r io.Reader, close func() error) ([]byte, error) {
	out := bytes.Buffer{}
	ch := make(chan error, 1)
	if close != nil {
		defer close() // nolint
	}

	go func() {
		buf := make([]byte, 8192)
		for c.Err() == nil {
			ct, err := r.Read(buf)
			if err != nil {
				ch <- err
				return
			}
			_, _ = out.Write(buf[0:ct])
			//log.Printf("read %d bytes...", ct)
		}
	}()
	select {
	case <-c.Done():
		return out.Bytes(), ctx.NewErrorf(c, "timeout: %w", c.Err())
	case err := <-ch:
		if err == io.EOF {
			return out.Bytes(), nil
		}
		return out.Bytes(), ctx.NewErrorf(c, "ReadAll: %w", err)
	}
}
