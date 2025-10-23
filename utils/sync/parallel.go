package sync

import (
	"sync"

	"github.com/ohait/forego/ctx"
)

func Go[T any](c ctx.C, items []T, max int, f func(c ctx.C, item T) error) error {
	limiter := make(chan struct{}, max)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	wg.Add(len(items))
	for _, item := range items {
		select {
		case limiter <- struct{}{}:
		case <-c.Done():
			wg.Done()
			return c.Err()
		}
		go func(item T) {
			defer wg.Done()
			defer func() { <-limiter }()
			err := f(c, item)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(item)
	}

	wg.Wait()
	return firstErr
}
