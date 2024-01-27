package caching

import (
	"container/list"
	"fmt"
	"sync"
)

type ResolverFunc[T any] func(string) (*T, error)
type UpdaterFunc[T any] func(*T, uint) error

type Cache[T any] struct {
	capacity uint
	backing  map[string]*list.Element
	lruList  *list.List

	resolver ResolverFunc[T]
	updater  UpdaterFunc[T]
	waitFor  uint

	mu sync.Mutex
}

type Page[T any] struct {
	key   string
	value *T
	hits  uint
}

func NewCache[T any](capacity uint, resolver ResolverFunc[T], updater UpdaterFunc[T]) *Cache[T] {
	return &Cache[T]{
		capacity: capacity,
		backing:  make(map[string]*list.Element),
		lruList:  list.New(),
		resolver: resolver,
		updater:  updater,
		waitFor:  10,
	}
}

func (c *Cache[T]) Resolve(key string) (*T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.backing[key]; ok {
		c.lruList.MoveToFront(el)
		page := el.Value.(*Page[T])

		page.hits += 1
		if page.hits >= c.waitFor {
			if err := c.updater(page.value, page.hits); err == nil {
				page.hits = 0
			}
		}

		return page.value, nil
	}

	if len(c.backing) >= int(c.capacity) {
		oldest := c.lruList.Back()
		if oldest != nil {
			page := oldest.Value.(*Page[T])
			_ = c.updater(page.value, page.hits)
			delete(c.backing, page.key)
			c.lruList.Remove(oldest)
		}
	}

	link, err := c.resolver(key)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %v: %v", key, err)
	}

	page := &Page[T]{key: key, value: link, hits: 1}
	c.backing[key] = c.lruList.PushFront(page)

	return link, nil
}
