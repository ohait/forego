package sync

import (
	"container/heap"
	"errors"
	"time"
)

// LRU is a thread-safe least-recently-used cache.
type LRU[K comparable, V any] struct {
	mu    Mutex
	cap   int
	items items[K, V]
	index map[K]*entry[K, V]
}

// NewLRU creates a new LRU cache with the given capacity.
func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	return &LRU[K, V]{
		cap:   capacity,
		index: make(map[K]*entry[K, V]),
	}
}

type items[K comparable, V any] []*entry[K, V]

var _ heap.Interface = (*items[string, any])(nil)

func (list items[K, V]) Len() int {
	return len(list)
}

func (list items[K, V]) Less(i, j int) bool {
	return list[i].last.Before(list[j].last)
}

func (list items[K, V]) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
	list[i].idx = i
	list[j].idx = j
}

func (list *items[K, V]) Push(x any) {
	ent := x.(*entry[K, V])
	ent.idx = len(*list)
	*list = append(*list, ent)
}

func (list *items[K, V]) Pop() any {
	old := *list
	n := len(old)
	x := old[n-1]
	old[n-1] = nil // avoid memory leak
	*list = old[0 : n-1]
	return x
}

type entry[K comparable, V any] struct {
	key   K
	value V
	error error
	last  time.Time
	idx   int
	mu    Mutex
}

// Memo returns the cached value for key, or calls f() to compute and cache it.
func (lru *LRU[K, V]) Memo(key K, f func() (V, error)) (value V, err error) {
	lru.mu.Lock()

	if ent, ok := lru.index[key]; ok {
		ent.last = time.Now()
		heap.Fix(&lru.items, ent.idx)
		lru.mu.Unlock()

		// acquire the lock to wait for the value to be ready
		ent.mu.Lock()
		defer ent.mu.Unlock()
		return ent.value, ent.error
	}

	// create the new entry, get a lock to it, release the lru lock and wait for the function
	ent := &entry[K, V]{
		key:   key,
		value: value,
		error: errors.New("panic"),
		last:  time.Now(),
	}
	heap.Push(&lru.items, ent)
	lru.index[key] = ent

	if len(lru.items) > lru.cap {
		oldest := heap.Pop(&lru.items).(*entry[K, V])
		delete(lru.index, oldest.key)
	}

	ent.mu.Lock()
	defer ent.mu.Unlock() // defer because f() might panic
	lru.mu.Unlock()

	ent.value, ent.error = f()
	return ent.value, ent.error
}

// Evict removes an entry from the cache if present.
func (lru *LRU[K, V]) Evict(key K) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	if ent, ok := lru.index[key]; ok {
		heap.Remove(&lru.items, ent.idx)
		delete(lru.index, key)
	}
}

// Len returns the number of entries in the cache.
func (lru *LRU[K, V]) Len() int {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	return len(lru.items)
}
