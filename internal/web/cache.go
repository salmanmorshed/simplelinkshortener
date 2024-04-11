package web

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

const (
	SyncInterval    = 10 * time.Second
	CleanupInterval = 30 * time.Second
)

type ResolveFunc func(string) (*db.Link, error)
type AugmentFunc func(*Page)
type CohereFunc func(*Page)

type Page struct {
	sync.RWMutex
	LinkID    uint
	LinkURL   string
	NewVisits uint
	lruMarker *list.Element
}

func (p *Page) isDirty() bool {
	p.RLock()
	defer p.RUnlock()
	return p.NewVisits > 0
}

type Cache struct {
	sync.RWMutex
	capacity uint
	backing  map[string]*Page
	lruList  *list.List
	cohereFn CohereFunc
}

func NewCache(capacity uint, cohere CohereFunc) *Cache {
	c := Cache{
		capacity: capacity,
		backing:  make(map[string]*Page),
		lruList:  list.New(),
		cohereFn: cohere,
	}

	coherenceTicker := time.NewTicker(SyncInterval)
	evictionTicker := time.NewTicker(CleanupInterval)

	go func() {
		for {
			select {
			case <-coherenceTicker.C:
				c.syncAllPages()
			case <-evictionTicker.C:
				c.evictOldPages()
			}
		}
	}()

	return &c
}

func NewCacheContext(ctx context.Context, capacity uint, cohere CohereFunc) *Cache {
	c := NewCache(capacity, cohere)

	if wg, ok := ctx.Value("ExitWG").(*sync.WaitGroup); ok {
		wg.Add(1)
		go func() {
			<-ctx.Done()
			c.Close()
			wg.Done()
		}()
	}

	return c
}

func (c *Cache) Resolve(key string, resolve ResolveFunc, augment AugmentFunc) (string, error) {
	c.RLock()
	page, exists := c.backing[key]
	if exists && page != nil {
		page.Lock()
		augment(page)
		page.Unlock()
		return page.LinkURL, nil
	}
	c.RUnlock()

	link, err := resolve(key)
	if err != nil {
		return "", fmt.Errorf("failed to resolve link %v: %v", key, err)
	}

	c.Lock()
	c.backing[key] = &Page{
		LinkID:    link.ID,
		LinkURL:   link.URL,
		NewVisits: 1,
		lruMarker: c.lruList.PushBack(key),
	}
	c.Unlock()

	return link.URL, nil
}

func (c *Cache) syncAllPages() {
	c.RLock()
	for _, page := range c.backing {
		if page.isDirty() {
			page.Lock()
			c.cohereFn(page)
			page.Unlock()
		}
	}
	c.RUnlock()
}

func (c *Cache) evictOldPages() {
	c.Lock()
	removeCount := c.lruList.Len() - int(c.capacity)
	for range removeCount {
		el := c.lruList.Back()
		key := el.Value.(string)
		page := c.backing[key]
		c.cohereFn(page)
		delete(c.backing, key)
		c.lruList.Remove(el)
	}
	c.Unlock()
}

func (c *Cache) Close() {
	c.syncAllPages()
}
