package caching

import (
	"container/list"
	"fmt"
	"sync"
)

type ResolveFunc[T any] func(string) (*T, error)
type CohereFunc[T any] func(*T, uint) error

type Cache[T any] struct {
	capacity    uint
	backingMap  map[string]*list.Element
	lruList     *list.List
	resolveFn   ResolveFunc[T]
	cohereFn    CohereFunc[T]
	cohereDelay uint
	mutex       sync.Mutex
}

type Page[T any] struct {
	key   string
	value *T
	hits  uint
}

func NewCache[T any](
	capacity uint,
	resolver ResolveFunc[T],
	coherer CohereFunc[T],
	cohereDelay uint,
) *Cache[T] {
	return &Cache[T]{
		capacity:    capacity,
		backingMap:  make(map[string]*list.Element),
		lruList:     list.New(),
		resolveFn:   resolver,
		cohereFn:    coherer,
		cohereDelay: cohereDelay,
	}
}

func (c *Cache[T]) Resolve(key string) (*T, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if el, ok := c.backingMap[key]; ok {
		c.lruList.MoveToFront(el)
		page := el.Value.(*Page[T])

		page.hits += 1
		if page.hits >= c.cohereDelay {
			if err := c.cohereFn(page.value, page.hits); err == nil {
				page.hits = 0
			}
		}

		return page.value, nil
	}

	if len(c.backingMap) >= int(c.capacity) {
		oldest := c.lruList.Back()
		if oldest != nil {
			page := oldest.Value.(*Page[T])
			_ = c.cohereFn(page.value, page.hits)
			delete(c.backingMap, page.key)
			c.lruList.Remove(oldest)
		}
	}

	link, err := c.resolveFn(key)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %v: %v", key, err)
	}

	page := &Page[T]{key: key, value: link, hits: 1}
	c.backingMap[key] = c.lruList.PushFront(page)

	return link, nil
}
