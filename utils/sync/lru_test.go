package sync_test

import (
	"errors"
	"testing"

	"github.com/ohait/forego/test"
	"github.com/ohait/forego/utils/sync"
)

func TestLRUBasic(t *testing.T) {
	lru := sync.NewLRU[string, int](3)

	calls := 0
	get := func(key string, val int) int {
		v, err := lru.Memo(key, func() (int, error) {
			calls++
			return val, nil
		})
		test.NoError(t, err)
		return v
	}

	// first call computes
	test.EqualsGo(t, 1, get("a", 1))
	test.EqualsGo(t, 1, calls)

	// second call returns cached
	test.EqualsGo(t, 1, get("a", 99))
	test.EqualsGo(t, 1, calls)

	// add more entries
	test.EqualsGo(t, 2, get("b", 2))
	test.EqualsGo(t, 3, get("c", 3))
	test.EqualsGo(t, 3, calls)

	// all three still cached
	test.EqualsGo(t, 1, get("a", 99))
	test.EqualsGo(t, 2, get("b", 99))
	test.EqualsGo(t, 3, get("c", 99))
	test.EqualsGo(t, 3, calls)
}

func TestLRUEviction(t *testing.T) {
	lru := sync.NewLRU[string, int](2)

	calls := 0
	get := func(key string, val int) int {
		v, err := lru.Memo(key, func() (int, error) {
			calls++
			return val, nil
		})
		test.NoError(t, err)
		return v
	}

	get("a", 1)
	get("b", 2)
	test.EqualsGo(t, 2, calls)

	// access "a" to make it recently used
	get("a", 99)
	test.EqualsGo(t, 2, calls)

	// add "c", should evict "b" (oldest)
	get("c", 3)
	test.EqualsGo(t, 3, calls)

	// "a" still cached
	get("a", 99)
	test.EqualsGo(t, 3, calls)

	// "b" was evicted, needs recompute
	get("b", 20)
	test.EqualsGo(t, 4, calls)
}

func TestLRUEvictAndLen(t *testing.T) {
	lru := sync.NewLRU[string, int](3)

	get := func(key string, val int) {
		_, err := lru.Memo(key, func() (int, error) {
			return val, nil
		})
		test.NoError(t, err)
	}

	test.EqualsGo(t, 0, lru.Len())

	get("a", 1)
	get("b", 2)
	get("c", 3)
	test.EqualsGo(t, 3, lru.Len())

	// evict existing key
	lru.Evict("b")
	test.EqualsGo(t, 2, lru.Len())

	// evict non-existing key (no-op)
	lru.Evict("z")
	test.EqualsGo(t, 2, lru.Len())

	// "b" was evicted, will recompute
	calls := 0
	v, err := lru.Memo("b", func() (int, error) {
		calls++
		return 20, nil
	})
	test.NoError(t, err)
	test.EqualsGo(t, 20, v)
	test.EqualsGo(t, 1, calls)
	test.EqualsGo(t, 3, lru.Len())
}

func TestLRUError(t *testing.T) {
	lru := sync.NewLRU[string, int](2)

	testErr := errors.New("test error")
	calls := 0
	_, err := lru.Memo("fail", func() (int, error) {
		calls++
		return 0, testErr
	})
	test.Assert(t, errors.Is(err, testErr))
	test.EqualsGo(t, 1, calls)

	// errors are cached (negative caching)
	_, err = lru.Memo("fail", func() (int, error) {
		calls++
		return 42, nil
	})
	test.Assert(t, errors.Is(err, testErr))
	test.EqualsGo(t, 1, calls) // f() not called again
}
