package sync_test

import (
	"testing"
	"time"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/test"
	"github.com/ohait/forego/utils/sync"
)

func TestMulti(t *testing.T) {
	c := test.Context(t)
	t0 := time.Now()
	sync.Go(c, []int{0, 1, 2, 3, 4, 5, 6, 7}, 3, func(c ctx.C, i int) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	elapsed := time.Since(t0)
	test.Assert(t, elapsed >= 250*time.Millisecond)
	test.Assert(t, elapsed < 400*time.Millisecond)
}
