package utils_test

import (
	"io"
	"testing"
	"time"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/test"
	"github.com/ohait/forego/utils"
)

func TestReadAll(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	c := test.Context(t)

	t.Run("timeout", func(t *testing.T) {
		r, w := io.Pipe()
		t0 := time.Now()
		c, cf := ctx.WithTimeout(c, 50*time.Millisecond)
		defer cf()
		in, err := utils.ReadAll(c, r, r.Close)
		test.Error(t, err)
		test.Assert(t, time.Since(t0) > 50*time.Millisecond)
		test.Empty(t, in)
		w.Close()
	})
	t.Run("slow", func(t *testing.T) {
		r, w := io.Pipe()
		t0 := time.Now()
		c, cf := ctx.WithTimeout(c, time.Second)
		defer cf()

		go func() {
			time.Sleep(20 * time.Millisecond)
			_, _ = w.Write([]byte("foo"))
			time.Sleep(20 * time.Millisecond)
			_, _ = w.Write([]byte("bar"))
			time.Sleep(20 * time.Millisecond)
			w.Close()
		}()

		in, err := utils.ReadAll(c, r, r.Close)
		test.NoError(t, err)
		test.Assert(t, time.Since(t0) > 50*time.Millisecond)
		test.EqualsGo(t, "foobar", string(in))
	})
}
