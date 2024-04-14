package web

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

var (
	CacheWaitGroup    sync.WaitGroup
	CoherenceInterval = 10 * time.Second
)

type ResolveFunc func(context.Context, string) (*db.Link, error)
type CohereFunc func(*Page)

type Page struct {
	lruMarker *list.Element
	LinkID    uint
	LinkURL   string
	NewVisits uint
}

type Cache struct {
	capacity uint
	resolver ResolveFunc
	coherer  CohereFunc

	backing map[string]*Page
	lruList *list.List

	lookupCh chan cacheLookup
	evictCh  chan struct{}
}

type cacheLookup struct {
	key  string
	ctx  context.Context
	done chan cacheResult
}

type cacheResult struct {
	url string
	err error
}

func NewCacheContext(ctx context.Context, capacity uint, resolver ResolveFunc, coherer CohereFunc) *Cache {
	c := Cache{
		capacity: capacity,
		resolver: resolver,
		coherer:  coherer,
		backing:  make(map[string]*Page),
		lruList:  list.New(),
		lookupCh: make(chan cacheLookup),
		evictCh:  make(chan struct{}, capacity),
	}

	CacheWaitGroup.Add(1)

	coherenceTicker := time.NewTicker(CoherenceInterval)
	defer coherenceTicker.Stop()

	go func() {
		for {
			select {
			case lookup := <-c.lookupCh:
				c.handleLookup(lookup)

			case <-coherenceTicker.C:
				c.syncAllPages()

			case <-c.evictCh:
				c.evictOldPages()

			case <-ctx.Done():
				c.syncAllPages()
				CacheWaitGroup.Done()
				return
			}
		}
	}()

	return &c
}

func (c *Cache) Lookup(ctx context.Context, key string) (string, error) {
	lookup := cacheLookup{key, ctx, make(chan cacheResult)}
	c.lookupCh <- lookup
	result := <-lookup.done
	return result.url, result.err
}

func (c *Cache) handleLookup(lookup cacheLookup) {
	if page, exists := c.backing[lookup.key]; exists && page != nil {
		c.lruList.MoveToFront(page.lruMarker)
		page.NewVisits += 1
		lookup.done <- cacheResult{page.LinkURL, nil}
		close(lookup.done)
		return
	}

	link, err := c.resolver(lookup.ctx, lookup.key)
	if err != nil {
		lookup.done <- cacheResult{"", err}
		close(lookup.done)
		return
	}

	c.backing[lookup.key] = &Page{
		LinkID:    link.ID,
		LinkURL:   link.URL,
		NewVisits: 1,
		lruMarker: c.lruList.PushBack(lookup.key),
	}
	lookup.done <- cacheResult{link.URL, nil}
	close(lookup.done)
	c.evictCh <- struct{}{}
}

func (c *Cache) syncAllPages() {
	for _, page := range c.backing {
		c.coherer(page)
	}
}

func (c *Cache) evictOldPages() {
	excess := c.lruList.Len() - int(c.capacity)
	for range excess {
		el := c.lruList.Back()
		key := el.Value.(string)
		page := c.backing[key]
		c.coherer(page)
		delete(c.backing, key)
		c.lruList.Remove(el)
	}
}

func (c *Cache) Close() {
	c.syncAllPages()
}
