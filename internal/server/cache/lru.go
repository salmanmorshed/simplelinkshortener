package cache

import (
	"container/list"
	"fmt"
	"sync"
)

type Cacheable any
type ResolverFunc[T Cacheable] func(string) (*T, error)
type CohererFunc[T Cacheable] func(*T, uint) error

type Cache[T Cacheable] struct {
	capacity   uint
	backingMap map[string]*list.Element
	lruList    *list.List
	mutex      sync.Mutex

	resolver ResolverFunc[T]
	coherer  CohererFunc[T]

	// number of page hits before coherer is called
	cohereInterval uint
}

type Page[T Cacheable] struct {
	key   string
	value *T
	hits  uint
}

func New[T Cacheable](
	capacity uint,
	resolver ResolverFunc[T],
	coherer CohererFunc[T],
	cohereInterval uint,
) *Cache[T] {
	c := Cache[T]{
		capacity:       capacity,
		backingMap:     make(map[string]*list.Element),
		lruList:        list.New(),
		resolver:       resolver,
		coherer:        coherer,
		cohereInterval: cohereInterval,
	}
	return &c
}

func (c *Cache[T]) Resolve(key string) (*T, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if el, ok := c.backingMap[key]; ok {
		c.lruList.MoveToFront(el)
		page := el.Value.(*Page[T])

		page.hits += 1
		if page.hits >= c.cohereInterval {
			if err := c.coherer(page.value, page.hits); err == nil {
				page.hits = 0
			}
		}

		return page.value, nil
	}

	if len(c.backingMap) >= int(c.capacity) {
		oldest := c.lruList.Back()
		if oldest != nil {
			page := oldest.Value.(*Page[T])
			_ = c.coherer(page.value, page.hits)
			delete(c.backingMap, page.key)
			c.lruList.Remove(oldest)
		}
	}

	link, err := c.resolver(key)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %v: %v", key, err)
	}

	page := &Page[T]{key: key, value: link, hits: 1}
	c.backingMap[key] = c.lruList.PushFront(page)

	return link, nil
}
